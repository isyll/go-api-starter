package config

import "time"

type DBCredentials struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

type DatabaseConfig struct {
	Credentials      DBCredentials `yaml:"credentials"`
	AppCredentials   DBCredentials `yaml:"app_credentials"`
	AdminCredentials DBCredentials `yaml:"admin_credentials"`

	ConnectionPool ConnectionPoolConfig `yaml:"connection_pool"`

	AppPool *ConnectionPoolConfig `yaml:"app_pool"`

	AdminPool *ConnectionPoolConfig `yaml:"admin_pool"`

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
