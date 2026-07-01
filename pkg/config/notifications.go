package config

import (
	"time"
)

// NotificationsConfig holds deep-link URL templates and the FCM
// push-notification worker settings loaded from
// configs/notifications.yaml.
type NotificationsConfig struct {
	// DeepLinks maps event types to deep-link URL templates that the
	// mobile clients open when a push notification is tapped.
	DeepLinks struct {
		// BaseURL is prepended to every generated deep link.
		BaseURL string `yaml:"base_url"`
		// Templates maps event_type keys to URL path templates.
		Templates map[string]string `yaml:"templates"`
	} `yaml:"deep_links"`
	// EventLinks is a flat map of event_type -> deep-link template,
	// used when the full DeepLinks nesting is not needed.
	EventLinks map[string]string `yaml:"event_deep_links"`
	// Worker holds concurrency and retry settings for the Asynq
	// FCM push-notification processor.
	Worker struct {
		Concurrency int           `yaml:"concurrency"`
		RetryMax    int           `yaml:"retry_max"`
		RetryDelay  time.Duration `yaml:"retry_delay"`
	} `yaml:"worker"`
	// Batch controls how notifications are grouped before flushing
	// to the FCM API.
	Batch struct {
		MaxSize int `yaml:"max_size"`
		// FlushInterval is how long the batcher waits before sending
		// a batch that has not reached MaxSize.
		FlushInterval time.Duration `yaml:"flush_interval"`
	} `yaml:"batch"`
	// QuietHours suppresses push delivery during user-configured
	// do-not-disturb windows.
	QuietHours struct {
		// CheckEnabled toggles quiet-hours enforcement globally.
		CheckEnabled bool `yaml:"check_enabled"`
		// DefaultTimezone is used when the user has no explicit TZ set.
		DefaultTimezone string `yaml:"default_timezone"`
	} `yaml:"quiet_hours"`
}
