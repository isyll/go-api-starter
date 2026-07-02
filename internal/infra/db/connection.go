// Package db opens GORM connections to PostgreSQL per role.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/isyll/go-api-starter/pkg/config"
	"github.com/isyll/go-api-starter/pkg/logger"

	"github.com/isyll/go-api-starter/internal/persistence"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type ExtensionChecker struct {
	db     *sql.DB
	logger *logger.Logger
}

type Role string

const (
	RoleApp       Role = "app"
	RoleAdmin     Role = "admin"
	RoleMigration Role = "migration"
)

func InitDatabase(
	cfg *config.DatabaseConfig,
	role Role,
	l *logger.Logger,
) (*gorm.DB, error) {
	var (
		sqlDB *sql.DB
		db    *gorm.DB
		err   error
	)

	sqlDB, err = connectToPostgres(cfg, role, l)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err = initializeExtensions(sqlDB, l); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf(
			"failed to initialize extensions: %w",
			err,
		)
	}

	db, err = createGormConnection(cfg, l, sqlDB, role)
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf(
			"failed to create gorm connection: %w",
			err,
		)
	}

	if err := healthCheck(db, l); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	registerQueryCallbacks(db)
	registerRLSCallback(db, role)

	return db, nil
}

func credentialsFor(
	cfg *config.DatabaseConfig, role Role,
) (config.DBCredentials, error) {
	switch role {
	case RoleApp:
		if cfg.AppCredentials.User == "" {
			return config.DBCredentials{}, fmt.Errorf(
				"app role credentials not configured: " +
					"set DB_API_USER and DB_API_PASSWORD",
			)
		}
		return cfg.AppCredentials, nil
	case RoleAdmin:
		if cfg.AdminCredentials.User == "" {
			return config.DBCredentials{}, fmt.Errorf(
				"admin role credentials not configured: " +
					"set DB_WORKER_USER and DB_WORKER_PASSWORD",
			)
		}
		return cfg.AdminCredentials, nil
	default:
		return cfg.Credentials, nil
	}
}

func poolFor(
	cfg *config.DatabaseConfig,
	role Role,
) config.ConnectionPoolConfig {
	switch role {
	case RoleApp:
		if cfg.AppPool != nil {
			return *cfg.AppPool
		}
	case RoleAdmin:
		if cfg.AdminPool != nil {
			return *cfg.AdminPool
		}
	case RoleMigration:
		if cfg.MigratePool != nil {
			return *cfg.MigratePool
		}
	}
	return cfg.ConnectionPool
}

func connectToPostgres(
	cfg *config.DatabaseConfig,
	role Role,
	l *logger.Logger,
) (*sql.DB, error) {
	searchPath := "public,auth,notifications,audit"

	stmtTimeout := cfg.StatementTimeoutMs
	if stmtTimeout <= 0 {
		stmtTimeout = 5000
	}

	creds, err := credentialsFor(cfg, role)
	if err != nil {
		return nil, fmt.Errorf(
			"database role misconfiguration: %w", err,
		)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s options='-c statement_timeout=%d'",
		creds.Host,
		creds.Port,
		creds.User,
		creds.Password,
		creds.DBName,
		creds.SSLMode,
		searchPath,
		stmtTimeout,
	)

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("cannot ping database: %w", err)
	}

	l.Info(
		"Initial database connection established",
		"role", string(role),
		"user", creds.User,
	)
	return sqlDB, nil
}

func initializeExtensions(sqlDB *sql.DB, logx *logger.Logger) error {
	checker := &ExtensionChecker{db: sqlDB, logger: logx}

	requiredExtensions := []string{
		"postgis",
		"uuid-ossp",
		"pgcrypto",
	}

	optionalExtensions := []string{
		"pg_stat_statements",
	}

	for _, ext := range requiredExtensions {
		if err := checker.ensureExtension(ext, true); err != nil {
			return fmt.Errorf(
				"critical extension %s failed: %w",
				ext,
				err,
			)
		}
	}

	for _, ext := range optionalExtensions {
		if err := checker.ensureExtension(ext, false); err != nil {
			logx.Warn(
				fmt.Sprintf(
					"Optional extension %s not available: %v",
					ext,
					err,
				),
			)
		}
	}

	return nil
}

func (ec *ExtensionChecker) ensureExtension(
	name string,
	required bool,
) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = $1)`

	if err := ec.db.QueryRowContext(
		context.Background(),
		query,
		name,
	).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check extension %s: %w", name, err)
	}

	if exists {
		ec.logger.Info(
			fmt.Sprintf("Extension %s already installed", name),
		)
		return nil
	}

	schemaClause := ""
	if name == "postgis" || name == "postgis_topology" {
		schemaClause = " SCHEMA public"
	}

	createQuery := fmt.Sprintf(
		"CREATE EXTENSION IF NOT EXISTS %s%s",
		name,
		schemaClause,
	)
	if _, err := ec.db.ExecContext(
		context.Background(),
		createQuery,
	); err != nil {
		if required {
			return fmt.Errorf(
				"failed to create required extension %s: %w",
				name,
				err,
			)
		}
		return err
	}

	ec.logger.Info(
		fmt.Sprintf("Extension %s installed successfully", name),
	)
	return nil
}

func createGormConnection(
	cfg *config.DatabaseConfig,
	logx *logger.Logger,
	sqlDB *sql.DB,
	role Role,
) (*gorm.DB, error) {
	var gormLogLevel gormLogger.LogLevel
	switch cfg.LogLevel {
	case "debug":
		gormLogLevel = gormLogger.Info
	case "info":
		gormLogLevel = gormLogger.Warn
	default:
		gormLogLevel = gormLogger.Error
	}

	slowQueryThreshold := cfg.Monitoring.SlowQueryThreshold

	gormConfig := &gorm.Config{
		Logger: gormLogger.New(
			logx,
			gormLogger.Config{
				SlowThreshold:             slowQueryThreshold,
				LogLevel:                  gormLogLevel,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
				ParameterizedQueries:      true,
			},
		),
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), gormConfig)
	if err != nil {
		return nil, err
	}

	configurePool(sqlDB, poolFor(cfg, role), string(role), logx)
	return db, nil
}

func healthCheck(db *gorm.DB, logx *logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	var postgisVersion string
	if err := db.Raw(
		"SELECT PostGIS_Version()",
	).Scan(&postgisVersion).Error; err != nil {
		logx.Warn(fmt.Sprintf("PostGIS health check failed: %v", err))
	} else {
		logx.Info(fmt.Sprintf("PostGIS version: %s", postgisVersion))
	}

	logx.Info("Database health check passed")
	return nil
}

func configurePool(
	sqlDB *sql.DB,
	pool config.ConnectionPoolConfig,
	label string,
	logx *logger.Logger,
) {
	lifetime, err := time.ParseDuration(
		pool.ConnectionMaxLifetime,
	)
	if err != nil {
		logx.Printf(
			"Invalid connection_max_lifetime, using default 1h: %v",
			err,
		)
		lifetime = time.Hour
	}

	idleTime, err := time.ParseDuration(
		pool.ConnectionMaxIdleTime,
	)
	if err != nil {
		logx.Printf(
			"Invalid connection_max_idle_time, using default 30m: %v",
			err,
		)
		idleTime = 30 * time.Minute
	}

	sqlDB.SetMaxOpenConns(pool.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(pool.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(lifetime)
	sqlDB.SetConnMaxIdleTime(idleTime)

	logx.Info(
		fmt.Sprintf("%s pool configured - MaxOpen: %d, MaxIdle: %d",
			label, pool.MaxOpenConnections,
			pool.MaxIdleConnections),
	)
}

func GetDatabaseStats(db *gorm.DB) (map[string]any, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()

	result := map[string]any{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}

	var statStatementsCount int64
	err = db.Raw("SELECT count(*) FROM pg_stat_statements").
		Scan(&statStatementsCount).
		Error
	if err == nil {
		result["pg_stat_statements_queries"] = statStatementsCount
	}

	return result, err
}

func registerQueryCallbacks(db *gorm.DB) {
	inc := func(d *gorm.DB) {
		if d.Statement == nil || d.Statement.Context == nil {
			return
		}
		persistence.IncrQueryCounter(d.Statement.Context)
	}
	cb := db.Callback()
	must := func(err error) {
		if err != nil {
			panic(fmt.Errorf(
				"register gorm callback: %w", err,
			))
		}
	}
	must(cb.Query().Before("gorm:query").Register(
		"db:count_query", inc,
	))
	must(cb.Create().Before("gorm:create").Register(
		"db:count_create", inc,
	))
	must(cb.Update().Before("gorm:update").Register(
		"db:count_update", inc,
	))
	must(cb.Delete().Before("gorm:delete").Register(
		"db:count_delete", inc,
	))
	must(cb.Raw().Before("gorm:raw").Register(
		"db:count_raw", inc,
	))
}
