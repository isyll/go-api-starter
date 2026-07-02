package config

import "time"

type SecurityConfig struct {
	Auth struct {
		MaxInactivityTimeout time.Duration `yaml:"max_inactivity_timeout"`
		MaxSessionAge        time.Duration `yaml:"max_session_age"`
		MaxDevicesPerUser    int           `yaml:"max_devices_per_user"`

		OAT struct {
			AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`
			RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
		} `yaml:"oat"`
	} `yaml:"auth"`

	IDObfuscation struct {
		MinLength int    `yaml:"min_length"`
		Alphabet  string `yaml:"alphabet"`
	} `yaml:"id_obfuscation"`
}
