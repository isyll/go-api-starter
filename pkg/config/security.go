package config

import "time"

// SecurityConfig holds authentication and ID-obfuscation settings
// loaded from configs/security.yaml.
type SecurityConfig struct {
	// Auth holds session and opaque-access-token lifetime rules.
	Auth struct {
		// MaxInactivityTimeout is how long a session may stay idle
		// before it is revoked.
		MaxInactivityTimeout time.Duration `yaml:"max_inactivity_timeout"`
		// MaxSessionAge is the hard upper bound on a session lifetime.
		MaxSessionAge time.Duration `yaml:"max_session_age"`
		// MaxDevicesPerUser caps simultaneous active sessions.
		MaxDevicesPerUser int `yaml:"max_devices_per_user"`

		// OAT holds access and refresh token expiry.
		OAT struct {
			AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`
			RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
		} `yaml:"oat"`
	} `yaml:"auth"`

	// IDObfuscation configures the Sqids encoder that turns internal
	// int64 IDs into short opaque strings in API responses.
	IDObfuscation struct {
		MinLength int    `yaml:"min_length"`
		Alphabet  string `yaml:"alphabet"`
	} `yaml:"id_obfuscation"`
}
