package config

import "time"

type SecurityConfig struct {
	Auth struct {
		MaxInactivityTimeout time.Duration `yaml:"max_inactivity_timeout"`
		MaxDevicesPerUser    int           `yaml:"max_devices_per_user"`

		OAT struct {
			AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`
			RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
		} `yaml:"oat"`

		// Lockout throttles password guessing per account. Zero values fall
		// back to 10 attempts in a 15 minute window.
		Lockout struct {
			MaxAttempts int           `yaml:"max_attempts"`
			Window      time.Duration `yaml:"window"`
		} `yaml:"lockout"`
	} `yaml:"auth"`

	IDObfuscation struct {
		MinLength int    `yaml:"min_length"`
		Alphabet  string `yaml:"alphabet"`
	} `yaml:"id_obfuscation"`

	// PasswordHash holds the argon2id cost parameters. Zero values fall back
	// to sensible defaults in the auth package.
	PasswordHash struct {
		Memory      uint32 `yaml:"memory"`
		Iterations  uint32 `yaml:"iterations"`
		Parallelism uint8  `yaml:"parallelism"`
		SaltLength  uint32 `yaml:"salt_length"`
		KeyLength   uint32 `yaml:"key_length"`
	} `yaml:"password_hash"`
}
