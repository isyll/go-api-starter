package config

import "time"

type AppConfig struct {
	Info struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"app"`

	Server struct {
		Port string `yaml:"port"`

		// Disable on internet-facing servers.
		Reflection bool `yaml:"reflection"`

		// Zero values keep the grpc-go defaults.
		MaxRecvMsgSizeBytes int           `yaml:"max_recv_msg_size_bytes"`
		MaxSendMsgSizeBytes int           `yaml:"max_send_msg_size_bytes"`
		ConnectionTimeout   time.Duration `yaml:"connection_timeout"`

		Keepalive struct {
			Time    time.Duration `yaml:"time"`
			Timeout time.Duration `yaml:"timeout"`

			MaxConnectionIdle     time.Duration `yaml:"max_connection_idle"`
			MaxConnectionAge      time.Duration `yaml:"max_connection_age"`
			MaxConnectionAgeGrace time.Duration `yaml:"max_connection_age_grace"`

			MinClientInterval time.Duration `yaml:"min_client_interval"`
		} `yaml:"keepalive"`

		ShutdownGrace time.Duration `yaml:"shutdown_grace"`
	} `yaml:"server"`

	I18n struct {
		DefaultLanguage string `yaml:"default_language"`
		LocalesDir      string `yaml:"locales_dir"`
	} `yaml:"i18n"`
}

func (c *AppConfig) GetServerAddress() string {
	return "0.0.0.0:" + c.Server.Port
}
