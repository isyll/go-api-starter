package config

import "time"

type EventsConfig struct {
	Outbox struct {
		Interval time.Duration `yaml:"interval"`

		MetricsInterval time.Duration `yaml:"metrics_interval"`

		DrainOnAPI bool `yaml:"drain_on_api"`
	} `yaml:"outbox"`
}

// applyDefaults fills the outbox intervals when they are left unset.
func (c *EventsConfig) applyDefaults() {
	if c.Outbox.Interval == 0 {
		c.Outbox.Interval = 5 * time.Second
	}
	if c.Outbox.MetricsInterval == 0 {
		c.Outbox.MetricsInterval = 15 * time.Second
	}
}
