package config

import "time"

// DBCredentials is one (user, password, db) tuple. The DatabaseConfig
// carries three of these to keep the migration / app / admin roles
// explicit and impossible to swap by accident.
type DBCredentials struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

// DatabaseConfig owns three logical PG roles:
//
//   - Credentials       — the migration role (full DDL, table ownership).
//     Used only by cmd/migrate.
//   - AppCredentials    — the API runtime role. Subject to row-level
//     security; every transaction runs SET LOCAL
//     app.current_user_id via the GORM callback.
//   - AdminCredentials  — workers and admin paths. Granted explicit
//     RLS-allow policies; the policy itself is the
//     audit trail.
//
// Production deployments override each tuple via env vars; defaults
// in configs/database.yaml are dev-only.
type DatabaseConfig struct {
	Credentials      DBCredentials `yaml:"credentials"`
	AppCredentials   DBCredentials `yaml:"app_credentials"`
	AdminCredentials DBCredentials `yaml:"admin_credentials"`

	// ConnectionPool is the global pool config used by all roles unless
	// a per-role override is set (AppPool / AdminPool / MigratePool).
	ConnectionPool ConnectionPoolConfig `yaml:"connection_pool"`

	// AppPool overrides ConnectionPool for the API runtime role
	// (app_api). The API serves short-lived user requests; a tighter
	// max-open keeps memory bounded under burst traffic. Falls back to
	// ConnectionPool when nil.
	AppPool *ConnectionPoolConfig `yaml:"app_pool"`

	// AdminPool overrides ConnectionPool for the workers/admin role
	// (app_worker). Workers run longer queries (event drain,
	// bulk jobs) may spike concurrency differently from the
	// API path. Falls back to ConnectionPool when nil.
	AdminPool *ConnectionPoolConfig `yaml:"admin_pool"`

	// MigratePool overrides ConnectionPool for the DDL role (app_owner).
	// Migrations are single-threaded and short-lived; a small pool is
	// sufficient. Falls back to ConnectionPool when nil.
	MigratePool *ConnectionPoolConfig `yaml:"migrate_pool"`

	StatementTimeoutMs int `yaml:"statement_timeout_ms"`

	Backup struct {
		Enabled       bool   `yaml:"enabled"`
		Schedule      string `yaml:"schedule"`
		RetentionDays int    `yaml:"retention_days"`
	} `yaml:"backup"`

	Monitoring struct {
		SlowQueryThreshold time.Duration `yaml:"slow_query_threshold"`
		LogQueries         bool          `yaml:"log_queries"`
		MetricsEnabled     bool          `yaml:"metrics_enabled"`
	} `yaml:"monitoring"`

	LogLevel string `yaml:"log_level"`
}
