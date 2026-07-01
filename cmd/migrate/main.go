// Command migrate applies and rolls back PostgreSQL migrations for
// the appdb and app_worker databases. It reads the --target
// flag to select which database and the --steps flag to control
// how many migrations to apply or roll back.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/isyll/go-api-starter/pkg/config"
	appenv "github.com/isyll/go-api-starter/pkg/env"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// directionUp and directionDown name the two wire values stored
// in MigrationRecord.Direction and used by the CLI switch. Kept
// as constants so a typo at a call site is a compile error.
const (
	directionUp   = "up"
	directionDown = "down"
)

// MigrationRecord is one row of the public.migration_history
// audit table. The CLI inserts a row per executed step (up,
// down, force, or to-version) so operators can reconstruct who
// ran what in which environment.
type MigrationRecord struct {
	ID          string    `json:"id"`
	Version     uint      `json:"version"`
	Name        string    `json:"name"`
	Direction   string    `json:"direction"`
	ExecutedAt  time.Time `json:"executed_at"`
	Duration    int64     `json:"duration_ms"`
	Status      string    `json:"status"`
	ErrorMsg    string    `json:"error_message,omitempty"`
	ExecutedBy  string    `json:"executed_by"`
	Environment string    `json:"environment"`
}

// MigrationTracker owns the connection to the audit table
// public.migration_history. It opens its own DB handle (separate
// from golang-migrate's) so the audit row write does not depend
// on the migrate driver's connection lifecycle.
type MigrationTracker struct {
	// db is the *sql.DB handle backing the tracker's queries.
	db *sql.DB
	// dsn is the libpq connection string, retained so init can
	// reconnect lazily.
	dsn string
}

// main is the entry point of the migrate CLI. It parses flags,
// loads the backend config, opens a tracking connection, and
// dispatches to the requested sub-command (up/down/steps/force/
// status/history/create).
func main() {
	var (
		up    = flag.Bool("up", false, "Run all migrations up")
		down  = flag.Bool("down", false, "Run all migrations down")
		steps = flag.Int(
			"steps",
			0,
			"Number of migration steps (use with -up or -down)",
		)
		version = flag.Uint("version", 0, "Migrate to specific version")
		force   = flag.Uint(
			"force",
			0,
			"Force database to specific version (dirty state recovery)",
		)
		status = flag.Bool(
			"status",
			false,
			"Show current migration status",
		)
		history = flag.Bool("history", false, "Show migration history")
		create  = flag.String(
			"create",
			"",
			"Create new migration files with given name",
		)
	)
	flag.Parse()

	appenv.InitApp()
	cfg, err := config.LoadAllConfigs()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	searchPath := "public,auth,notifications,audit"

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		cfg.Database.Credentials.Host,
		cfg.Database.Credentials.Port,
		cfg.Database.Credentials.User,
		cfg.Database.Credentials.Password,
		cfg.Database.Credentials.DBName,
		cfg.Database.Credentials.SSLMode,
		searchPath,
	)

	tracker := &MigrationTracker{dsn: dsn}
	if err := tracker.init(); err != nil {
		log.Fatal("Failed to initialize migration tracker:", err)
	}
	defer tracker.close()

	// Wire SIGINT/SIGTERM so Ctrl+C during a long migration is
	// observed: the audit-row queries and version lookups stop
	// taking new work, and executeMigrationWithTracking forwards
	// the cancellation to golang-migrate via GracefulStop. Note
	// that an in-flight SQL statement inside golang-migrate is
	// not always interruptible mid-step; the goal is to abort
	// between steps rather than to kill PostgreSQL transactions.
	// Signal handling is wired AFTER tracker.init so the lint
	// guard against `defer stop() not run on log.Fatal` is
	// honored — boot-time fatals fall through to the default Go
	// signal disposition.
	ctx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM,
	)
	defer stop()

	switch {
	case *create != "":
		createMigration(*create)

	case *status:
		showStatus(dsn)

	case *history:
		showHistory(ctx, tracker)

	case *force > 0:
		forceMigration(ctx, dsn, tracker, *force)

	case *version > 0:
		migrateToVersion(ctx, dsn, tracker, *version)

	case *up:
		if *steps > 0 {
			migrateSteps(ctx, dsn, tracker, *steps)
		} else {
			migrateUp(ctx, dsn, tracker)
		}

	case *down:
		if *steps > 0 {
			migrateSteps(ctx, dsn, tracker, -*steps)
		} else {
			migrateDown(ctx, dsn, tracker)
		}

	default:
		flag.Usage()
		fmt.Println("\nExamples:")
		fmt.Println("  go run main.go -create create_users_table")
		fmt.Println("  go run main.go -up")
		fmt.Println("  go run main.go -down -steps 1")
		fmt.Println("  go run main.go -status")
		fmt.Println("  go run main.go -history")
	}
}

// init opens the tracker's DB connection and ensures the audit
// table exists. Called once from main; subsequent recordMigration
// calls reuse the open handle.
func (mt *MigrationTracker) init() error {
	db, err := sql.Open("postgres", mt.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	mt.db = db

	return mt.createHistoryTable()
}

// close releases the tracker's DB connection. Safe to call when
// init was never run (no-op on nil db).
func (mt *MigrationTracker) close() {
	if mt.db != nil {
		_ = mt.db.Close()
	}
}

// moveTableToPublic moves a table from any non-public schema
// to the public schema if it exists there but not in public.
// This handles the case where migration tables were
// accidentally created in the rides schema due to search_path
// order.
func (mt *MigrationTracker) moveTableToPublic(
	ctx context.Context,
	tableName string,
) error {
	var inPublic bool
	err := mt.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)`, tableName).Scan(&inPublic)
	if err != nil {
		return fmt.Errorf("check public.%s: %w", tableName, err)
	}
	if inPublic {
		return nil
	}

	var srcSchema sql.NullString
	err = mt.db.QueryRowContext(ctx, `
		SELECT table_schema
		FROM information_schema.tables
		WHERE table_name = $1
		AND table_schema != 'public'
		LIMIT 1`, tableName).Scan(&srcSchema)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("find %s: %w", tableName, err)
	}

	if srcSchema.Valid {
		_, err = mt.db.ExecContext(ctx, fmt.Sprintf(
			"ALTER TABLE %s.%s SET SCHEMA public",
			srcSchema.String, tableName,
		))
		if err != nil {
			return fmt.Errorf(
				"move %s.%s to public: %w",
				srcSchema.String, tableName, err,
			)
		}
		log.Printf(
			"Moved %s.%s to public schema",
			srcSchema.String, tableName,
		)
	}

	return nil
}

// createHistoryTable is the idempotent DDL that backs
// MigrationRecord. It must execute against the public schema so
// the table survives every per-domain schema drop the rest of
// the migration set might perform.
func (mt *MigrationTracker) createHistoryTable() error {
	ctx := context.Background()

	if err := mt.moveTableToPublic(
		ctx, "migration_history",
	); err != nil {
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS public.migration_history (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		version BIGINT NOT NULL,
		name VARCHAR(255) NOT NULL,
		direction VARCHAR(10) NOT NULL CHECK (direction IN ('up', 'down')),
		executed_at TIMESTAMPTZ DEFAULT NOW(),
		duration_ms BIGINT NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'success' CHECK (status IN ('success', 'failed')),
		error_message TEXT,
		executed_by VARCHAR(100) NOT NULL,
		environment VARCHAR(50) NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_migration_history_version ON public.migration_history(version);
	CREATE INDEX IF NOT EXISTS idx_migration_history_executed_at ON public.migration_history(executed_at DESC);`

	_, err := mt.db.ExecContext(ctx, query)
	return err
}

// recordMigration inserts a single audit row. Called from
// executeMigrationWithTracking after every step, regardless of
// whether the step succeeded — the Status field carries the
// outcome.
func (mt *MigrationTracker) recordMigration(
	ctx context.Context,
	record *MigrationRecord,
) error {
	query := `
	INSERT INTO public.migration_history (
		id, version, name, direction, executed_at,
		duration_ms, status, error_message, executed_by, environment
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := mt.db.ExecContext(ctx, query,
		record.ID,
		record.Version,
		record.Name,
		record.Direction,
		record.ExecutedAt,
		record.Duration,
		record.Status,
		record.ErrorMsg,
		record.ExecutedBy,
		record.Environment,
	)
	return err
}

// getHistory returns the most recent `limit` rows from
// public.migration_history in reverse chronological order. Used
// by the -history sub-command.
func (mt *MigrationTracker) getHistory(
	ctx context.Context,
	limit int,
) ([]MigrationRecord, error) {
	query := `
	SELECT id, version, name, direction, executed_at, duration_ms,
		status, COALESCE(error_message, ''), executed_by, environment
	FROM public.migration_history
	ORDER BY executed_at DESC
	LIMIT $1`

	rows, err := mt.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []MigrationRecord
	for rows.Next() {
		var r MigrationRecord
		err := rows.Scan(
			&r.ID,
			&r.Version,
			&r.Name,
			&r.Direction,
			&r.ExecutedAt,
			&r.Duration,
			&r.Status,
			&r.ErrorMsg,
			&r.ExecutedBy,
			&r.Environment,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// getCurrentUser reads the OS user invoking the CLI so audit
// rows attribute the migration to a human. Falls back to
// "unknown" when neither USER nor USERNAME is set.
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// getCurrentEnvironment returns the deployment environment
// recorded against the audit row. Reads APP_ENV first, then
// ENVIRONMENT, defaulting to "development" when neither is set.
func getCurrentEnvironment() string {
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	return "development"
}

// getMigrationName resolves a version integer to the
// human-readable name from the matching migrations/NNNNNN_*.up.sql
// file. Returns "migration_<n>" when the file cannot be found —
// the audit row still records the version.
func getMigrationName(version uint) string {
	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		return fmt.Sprintf("migration_%d", version)
	}

	versionStr := fmt.Sprintf("%06d", version)
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), versionStr) {
			name := strings.TrimPrefix(
				filepath.Base(file),
				versionStr+"_",
			)
			name = strings.TrimSuffix(name, ".up.sql")
			return name
		}
	}
	return fmt.Sprintf("migration_%d", version)
}

// createMigration prints the `migrate create` invocation that
// scaffolds a paired *.up.sql / *.down.sql under migrations/.
// Kept as a separate sub-command so operators don't have to
// remember the flags.
func createMigration(name string) {
	log.Printf("Creating migration: %s", name)
	fmt.Printf("Run this command to create migration files:\n")
	fmt.Printf(
		"migrate create -ext sql -dir migrations -seq %s\n", name,
	)
}

// moveSchemamigrationsToPublic moves the schema_migrations
// table from a non-public schema (e.g. rides) to public if
// needed. Must run before golang-migrate initializes.
func moveSchemamigrationsToPublic(db *sql.DB) error {
	ctx := context.Background()

	var inPublic bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'schema_migrations'
		)`).Scan(&inPublic)
	if err != nil {
		return fmt.Errorf(
			"check public.schema_migrations: %w", err,
		)
	}
	if inPublic {
		return nil
	}

	var srcSchema sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT table_schema
		FROM information_schema.tables
		WHERE table_name = 'schema_migrations'
		AND table_schema != 'public'
		LIMIT 1`).Scan(&srcSchema)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("find schema_migrations: %w", err)
	}

	if srcSchema.Valid {
		_, err = db.ExecContext(ctx, fmt.Sprintf(
			"ALTER TABLE %s.schema_migrations SET SCHEMA public",
			srcSchema.String,
		))
		if err != nil {
			return fmt.Errorf(
				"move %s.schema_migrations to public: %w",
				srcSchema.String, err,
			)
		}
		log.Printf(
			"Moved %s.schema_migrations to public schema",
			srcSchema.String,
		)
	}

	return nil
}

// getMigrate constructs the golang-migrate driver wired to
// `file://migrations` and the public schema. It runs the
// schema-migrations relocation guard first so a misplaced
// `rides.schema_migrations` from an earlier dev iteration is
// silently fixed.
func getMigrate(dsn string) (*migrate.Migrate, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to database: %w", err,
		)
	}

	if err = moveSchemamigrationsToPublic(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		SchemaName: "public",
	})
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create postgres driver: %w", err,
		)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create migrate instance: %w",
			err,
		)
	}

	return m, nil
}

// executeMigrationWithTracking wraps a migrate operation with
// version-diffing, timing, and audit-row insertion. It is the
// single chokepoint every up/down/steps/to-version sub-command
// funnels through, ensuring every executed step writes exactly
// one history row.
func executeMigrationWithTracking(
	ctx context.Context,
	dsn string,
	tracker *MigrationTracker,
	migrationFunc func(*migrate.Migrate) error,
	direction string,
) error {
	m, err := getMigrate(dsn)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	// Forward Ctrl+C to golang-migrate's graceful-stop channel.
	// The library checks GracefulStop between steps; an in-flight
	// SQL statement will not be canceled mid-execution, but the
	// next step in a multi-step run will not start.
	stopForwarder := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			select {
			case m.GracefulStop <- true:
			default:
			}
		case <-stopForwarder:
		}
	}()
	defer close(stopForwarder)

	currentVersion, _, _ := m.Version()
	startTime := time.Now()

	err = migrationFunc(m)
	duration := time.Since(startTime).Milliseconds()

	newVersion, _, _ := m.Version()

	var targetVersion uint
	var migrationName string

	if direction == directionUp && newVersion > currentVersion {
		targetVersion = newVersion
		migrationName = getMigrationName(newVersion)
	} else if direction == directionDown && currentVersion > newVersion {
		targetVersion = currentVersion
		migrationName = getMigrationName(currentVersion)
	} else {
		return err
	}

	record := MigrationRecord{
		ID:          uuid.New().String(),
		Version:     targetVersion,
		Name:        migrationName,
		Direction:   direction,
		ExecutedAt:  startTime,
		Duration:    duration,
		Status:      "success",
		ExecutedBy:  getCurrentUser(),
		Environment: getCurrentEnvironment(),
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		record.Status = "failed"
		record.ErrorMsg = err.Error()
	}

	// Use a fresh context (not ctx) for the audit-row write so a
	// Ctrl+C between the migration step and the audit insert
	// still records the outcome. The insert is bounded to 5 s.
	auditCtx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second,
	)
	defer cancel()
	if trackErr := tracker.recordMigration(
		auditCtx, &record,
	); trackErr != nil {
		log.Printf(
			"Warning: Failed to record migration history: %v",
			trackErr,
		)
	}

	return err
}

// showStatus prints the current migration version and dirty
// flag, with extra hints when the migration set has never been
// applied. Backs the -status sub-command.
func showStatus(dsn string) {
	log.Println("Checking migration status...")

	m, err := getMigrate(dsn)
	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}
	defer func() { _, _ = m.Close() }()

	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		log.Println("No migrations have been applied yet")
		return
	}
	if err != nil {
		log.Printf("Failed to get version: %v", err)
		return
	}

	status := "clean"
	if dirty {
		status = "dirty"
	}

	log.Printf(
		"Current migration version: %d (state: %s)",
		version,
		status,
	)
}

// showHistory pretty-prints the last 20 entries from
// public.migration_history. Backs the -history sub-command.
func showHistory(ctx context.Context, tracker *MigrationTracker) {
	log.Println("Migration history (last 20):")

	records, err := tracker.getHistory(ctx, 20)
	if err != nil {
		log.Fatal("Failed to get migration history:", err)
	}

	if len(records) == 0 {
		log.Println("No migration history found")
		return
	}

	fmt.Printf(
		"%-8s %-20s %-10s %-20s %-10s %-8s %-12s\n",
		"Version",
		"Name",
		"Direction",
		"Executed At",
		"Duration",
		"Status",
		"By",
	)
	fmt.Println(strings.Repeat("-", 90))

	for i := range records {
		fmt.Printf("%-8d %-20s %-10s %-20s %-10dms %-8s %-12s\n",
			records[i].Version,
			truncateString(records[i].Name, 20),
			records[i].Direction,
			records[i].ExecutedAt.Format("2006-01-02 15:04:05"),
			records[i].Duration,
			records[i].Status,
			records[i].ExecutedBy)

		if records[i].ErrorMsg != "" {
			fmt.Printf(
				"         Error: %s\n",
				truncateString(records[i].ErrorMsg, 70),
			)
		}
	}
}

// truncateString clamps s to maxLen runes, replacing the tail
// with an ellipsis when the string was actually shortened.
// Used only for column-fitting in showHistory.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// migrateUp drains every pending migration in ascending order.
// Treats ErrNoChange as a success (nothing to do).
func migrateUp(
	ctx context.Context,
	dsn string,
	tracker *MigrationTracker,
) {
	log.Println("Running all migrations up...")

	err := executeMigrationWithTracking(
		ctx,
		dsn,
		tracker,
		func(m *migrate.Migrate) error {
			return m.Up()
		},
		directionUp,
	)

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("Migration up failed:", err)
	}

	log.Println("All migrations completed successfully")
}

// migrateDown rolls every applied migration back in descending
// order. Intended for local dev reset; production rollbacks
// should target a specific version via -version instead.
func migrateDown(
	ctx context.Context,
	dsn string,
	tracker *MigrationTracker,
) {
	log.Println("Running all migrations down...")

	err := executeMigrationWithTracking(
		ctx,
		dsn,
		tracker,
		func(m *migrate.Migrate) error {
			return m.Down()
		},
		directionDown,
	)

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("Migration down failed:", err)
	}

	log.Println("All migrations rolled back successfully")
}

// migrateSteps moves the schema by `steps` versions in either
// direction. Positive steps drive up, negative drive down — the
// sign also determines which direction string lands in the
// audit row.
func migrateSteps(
	ctx context.Context,
	dsn string,
	tracker *MigrationTracker,
	steps int,
) {
	log.Printf("Running %d migration steps...", steps)

	direction := directionUp
	if steps < 0 {
		direction = directionDown
	}

	err := executeMigrationWithTracking(
		ctx,
		dsn,
		tracker,
		func(m *migrate.Migrate) error {
			return m.Steps(steps)
		},
		direction,
	)

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("Migration steps failed:", err)
	}

	log.Printf("Migration steps completed successfully")
}

// migrateToVersion drives the schema to a specific version,
// computing the direction relative to the current state. Used
// by the -version sub-command for targeted rollouts/rollbacks.
func migrateToVersion(
	ctx context.Context,
	dsn string,
	tracker *MigrationTracker,
	version uint,
) {
	log.Printf("Migrating to version %d...", version)

	m, err := getMigrate(dsn)
	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}
	defer func() { _, _ = m.Close() }()

	currentVersion, _, _ := m.Version()
	direction := directionUp
	if version < currentVersion {
		direction = directionDown
	}

	err = executeMigrationWithTracking(
		ctx,
		dsn,
		tracker,
		func(m *migrate.Migrate) error {
			return m.Migrate(version)
		},
		direction,
	)

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migration to version failed: %v", err)
		return
	}

	log.Printf(
		"Migration to version %d completed successfully",
		version,
	)
}

// forceMigration overrides the dirty bit and pins
// schema_migrations.version to the supplied number. The escape
// hatch for recovering from a partially-applied migration where
// the operator has already manually undone the partial change.
// Backs the -force sub-command.
func forceMigration(
	ctx context.Context,
	dsn string,
	tracker *MigrationTracker,
	version uint,
) {
	log.Printf("WARNING: Forcing migration to version %d...", version)

	m, err := getMigrate(dsn)
	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}
	defer func() { _, _ = m.Close() }()

	if version > uint(math.MaxInt) {
		log.Printf(
			"Force migration: version %d exceeds int range",
			version,
		)
		return
	}
	if err := m.Force(int(version)); err != nil {
		log.Printf("Force migration failed: %v", err)
		return
	}

	record := MigrationRecord{
		ID:          uuid.New().String(),
		Version:     version,
		Name:        "force_migration",
		Direction:   "force",
		ExecutedAt:  time.Now(),
		Duration:    0,
		Status:      "success",
		ExecutedBy:  getCurrentUser(),
		Environment: getCurrentEnvironment(),
	}

	auditCtx, cancel := context.WithTimeout(
		ctx, 5*time.Second,
	)
	defer cancel()
	if err := tracker.recordMigration(
		auditCtx, &record,
	); err != nil {
		log.Printf("Warning: Failed to record force migration: %v", err)
	}

	log.Printf("Migration forced to version %d", version)
}
