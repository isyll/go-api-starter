package config

import "time"

// AppConfig holds the server settings, i18n locale paths, and
// pagination defaults loaded from configs/api.yaml.
type AppConfig struct {
	// Info is the application name and version, surfaced in health.
	Info struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"app"`

	// Server is the gRPC bind port and timeouts.
	Server struct {
		Port            string        `yaml:"port"`
		ReadTimeout     time.Duration `yaml:"read_timeout"`
		WriteTimeout    time.Duration `yaml:"write_timeout"`
		RequestTimeout  time.Duration `yaml:"request_timeout"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
		IdleTimeout     time.Duration `yaml:"idle_timeout"`
	} `yaml:"server"`

	// I18n points to the locale directory (used by the email worker).
	I18n struct {
		DefaultLanguage string `yaml:"default_language"`
		LocalesDir      string `yaml:"locales_dir"`
	} `yaml:"i18n"`

	Pagination *PaginationConfig `yaml:"pagination"`
}

// GetServerAddress returns the "0.0.0.0:port" bind address.
func (c *AppConfig) GetServerAddress() string {
	return "0.0.0.0:" + c.Server.Port
}

// GetServerTimeouts returns read, write, and idle timeouts.
func (c *AppConfig) GetServerTimeouts() (read, write, idle time.Duration) {
	return c.Server.ReadTimeout, c.Server.WriteTimeout, c.Server.IdleTimeout
}
