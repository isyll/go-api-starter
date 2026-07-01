package config

import "time"

// EventsConfig governs the events package: outbox drain
// cadence, where the drain runs, and stampede backoff
// schedules. All values have sane defaults applied by
// LoadAllConfigs when the file is absent (events.yaml is
// optional — its presence is what changes behavior).
type EventsConfig struct {
	Outbox struct {
		// Interval is the periodic tick at which the drain
		// goroutine scans events.outbox for pending rows.
		// The worker drain runs at 5 s; the API fallback
		// drain runs at 30 s.
		Interval time.Duration `yaml:"interval"`
		// MetricsInterval is the cadence at which outbox
		// gauges are refreshed from the database.
		MetricsInterval time.Duration `yaml:"metrics_interval"`
		// DrainOnAPI controls whether the API binary runs
		// its own drain goroutine. Default false: the
		// canonical drain lives in the event-dispatcher
		// worker. Set to true only as a kill switch when
		// the worker is down or being debugged. Concurrent
		// drainers across replicas are safe (PendingBatch
		// uses FOR UPDATE SKIP LOCKED) — at worst they
		// waste a query per tick.
		DrainOnAPI bool `yaml:"drain_on_api"`
	} `yaml:"outbox"`
}

// EventsDefaults returns the EventsConfig used when no
// events.yaml is loaded. Centralized so the API and the
// event-dispatcher worker pick the same numbers.
func EventsDefaults() *EventsConfig {
	c := &EventsConfig{}
	c.Outbox.Interval = 5 * time.Second
	c.Outbox.MetricsInterval = 15 * time.Second
	c.Outbox.DrainOnAPI = false
	return c
}
