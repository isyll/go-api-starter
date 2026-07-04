package config

type DBCredentials struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type DatabaseConfig struct {
	Credentials      DBCredentials `yaml:"credentials"`
	AppCredentials   DBCredentials `yaml:"app_credentials"`
	AdminCredentials DBCredentials `yaml:"admin_credentials"`

	ConnectionPool ConnectionPoolConfig  `yaml:"connection_pool"`
	AppPool        *ConnectionPoolConfig `yaml:"app_pool"`
	AdminPool      *ConnectionPoolConfig `yaml:"admin_pool"`
	MigratePool    *ConnectionPoolConfig `yaml:"migrate_pool"`

	StatementTimeoutMs int `yaml:"statement_timeout_ms"`
}
