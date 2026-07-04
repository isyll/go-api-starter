package config

type ConnectionPoolConfig struct {
	MaxOpenConnections    int    `yaml:"max_open_connections"`
	ConnectionMaxLifetime string `yaml:"connection_max_lifetime"`
	ConnectionMaxIdleTime string `yaml:"connection_max_idle_time"`
}
