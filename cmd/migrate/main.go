// Command migrate applies and rolls back PostgreSQL migrations for
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

const (
	directionUp   = "up"
	directionDown = "down"
)

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

type MigrationTracker struct {
	db  *sql.DB
	dsn string
}

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

func (mt *MigrationTracker) init() error {
	db, err := sql.Open("postgres", mt.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	mt.db = db

	return mt.createHistoryTable()
}

func (mt *MigrationTracker) close() {
	if mt.db != nil {
		_ = mt.db.Close()
	}
}

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

func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

func getCurrentEnvironment() string {
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	return "development"
}

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

func createMigration(name string) {
	log.Printf("Creating migration: %s", name)
	fmt.Printf("Run this command to create migration files:\n")
	fmt.Printf(
		"migrate create -ext sql -dir migrations -seq %s\n", name,
	)
}

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

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

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
