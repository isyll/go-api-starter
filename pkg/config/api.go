package config

import "time"

type AppConfig struct {
	Info struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"app"`

	Server struct {
		Port            string        `yaml:"port"`
		ReadTimeout     time.Duration `yaml:"read_timeout"`
		WriteTimeout    time.Duration `yaml:"write_timeout"`
		RequestTimeout  time.Duration `yaml:"request_timeout"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
		IdleTimeout     time.Duration `yaml:"idle_timeout"`
	} `yaml:"server"`

	I18n struct {
		DefaultLanguage string `yaml:"default_language"`
		LocalesDir      string `yaml:"locales_dir"`
	} `yaml:"i18n"`

	Pagination *PaginationConfig `yaml:"pagination"`
}

func (c *AppConfig) GetServerAddress() string {
	return "0.0.0.0:" + c.Server.Port
}

func (c *AppConfig) GetServerTimeouts() (read, write, idle time.Duration) {
	return c.Server.ReadTimeout, c.Server.WriteTimeout, c.Server.IdleTimeout
}
