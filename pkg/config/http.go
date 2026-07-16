package config

import "time"

// Always-on HTTP surface: grpc-gateway mirror, CORS, webhooks.
type HTTPConfig struct {
	ListenAddr    string         `yaml:"listen_addr"`
	CORS          CORSConfig     `yaml:"cors"`
	Webhooks      WebhooksConfig `yaml:"webhooks"`
	ShutdownGrace time.Duration  `yaml:"shutdown_grace"`
}

// List fields are comma-separated so they survive env substitution.
type CORSConfig struct {
	AllowedOrigins   string        `yaml:"allowed_origins"`
	AllowedMethods   string        `yaml:"allowed_methods"`
	AllowedHeaders   string        `yaml:"allowed_headers"`
	AllowCredentials bool          `yaml:"allow_credentials"`
	MaxAge           time.Duration `yaml:"max_age"`
}

// Empty secrets reject every request for that provider.
type WebhooksConfig struct {
	Wave struct {
		Secret string `yaml:"secret"`
	} `yaml:"wave"`
	OrangeMoney struct {
		Secret string `yaml:"secret"`
		Header string `yaml:"header"`
	} `yaml:"orange_money"`
	PayDunya struct {
		MasterKey string `yaml:"master_key"`
	} `yaml:"paydunya"`
	Stripe struct {
		SigningSecret string        `yaml:"signing_secret"`
		Tolerance     time.Duration `yaml:"tolerance"`
	} `yaml:"stripe"`
}
