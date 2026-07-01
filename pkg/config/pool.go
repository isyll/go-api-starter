package config

// ConnectionPoolConfig holds the GORM/pgxpool connection-pool
// settings shared by the app and admin database configurations.
type ConnectionPoolConfig struct {
	// MaxOpenConnections is the cap on the number of open connections
	// in the pool at any time.
	MaxOpenConnections int `yaml:"max_open_connections"`
	// MaxIdleConnections is the number of connections the pool keeps
	// open when idle.
	MaxIdleConnections int `yaml:"max_idle_connections"`
	// ConnectionMaxLifetime is the maximum duration a single
	// connection may be reused (parsed as a Go duration string).
	ConnectionMaxLifetime string `yaml:"connection_max_lifetime"`
	// ConnectionMaxIdleTime is the maximum duration a connection may
	// remain idle in the pool before being closed.
	ConnectionMaxIdleTime string `yaml:"connection_max_idle_time"`
}
