package config

import "time"

type EventsConfig struct {
	Outbox struct {
		Interval time.Duration `yaml:"interval"`

		MetricsInterval time.Duration `yaml:"metrics_interval"`

		DrainOnAPI bool `yaml:"drain_on_api"`
	} `yaml:"outbox"`
}

func EventsDefaults() *EventsConfig {
	c := &EventsConfig{}
	c.Outbox.Interval = 5 * time.Second
	c.Outbox.MetricsInterval = 15 * time.Second
	c.Outbox.DrainOnAPI = false
	return c
}
