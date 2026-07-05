package config

import "time"

// MaintenanceConfig drives the worker's periodic retention sweeps.
type MaintenanceConfig struct {
	Retention struct {
		// Interval is how often the sweeper runs.
		Interval time.Duration `yaml:"interval"`
		// RefreshTokens removes expired or revoked refresh tokens older
		// than this age.
		RefreshTokens time.Duration `yaml:"refresh_tokens"`
		// ProcessedOutbox removes processed outbox rows older than this age.
		ProcessedOutbox time.Duration `yaml:"processed_outbox"`
		// LoginAttempts removes login attempt rows older than this age.
		LoginAttempts time.Duration `yaml:"login_attempts"`
	} `yaml:"retention"`
}

func (c *MaintenanceConfig) applyDefaults() {
	r := &c.Retention
	if r.Interval == 0 {
		r.Interval = time.Hour
	}
	if r.RefreshTokens == 0 {
		r.RefreshTokens = 30 * 24 * time.Hour
	}
	if r.ProcessedOutbox == 0 {
		r.ProcessedOutbox = 7 * 24 * time.Hour
	}
	if r.LoginAttempts == 0 {
		r.LoginAttempts = 90 * 24 * time.Hour
	}
}
