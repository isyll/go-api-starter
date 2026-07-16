package config

import "time"

type EventsConfig struct {
	Outbox struct {
		Interval time.Duration `yaml:"interval"`

		// Rows claimed per drain transaction; drain loops while batches stay full.
		BatchSize int `yaml:"batch_size"`

		MetricsInterval time.Duration `yaml:"metrics_interval"`

		DrainOnAPI bool `yaml:"drain_on_api"`
	} `yaml:"outbox"`

	Worker struct {
		Concurrency int `yaml:"concurrency"`
	} `yaml:"worker"`
}

func (c *EventsConfig) applyDefaults() {
	if c.Outbox.Interval == 0 {
		c.Outbox.Interval = 5 * time.Second
	}
	if c.Outbox.BatchSize == 0 {
		c.Outbox.BatchSize = 100
	}
	if c.Outbox.MetricsInterval == 0 {
		c.Outbox.MetricsInterval = 15 * time.Second
	}
	if c.Worker.Concurrency == 0 {
		c.Worker.Concurrency = 10
	}
}
