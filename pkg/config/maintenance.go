package config

import "time"

type MaintenanceConfig struct {
	Retention struct {
		Interval        time.Duration `yaml:"interval"`
		RefreshTokens   time.Duration `yaml:"refresh_tokens"`
		ProcessedOutbox time.Duration `yaml:"processed_outbox"`
		LoginAttempts   time.Duration `yaml:"login_attempts"`
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
