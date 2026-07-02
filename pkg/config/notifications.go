package config

import (
	"time"
)

type NotificationsConfig struct {
	DeepLinks struct {
		BaseURL   string            `yaml:"base_url"`
		Templates map[string]string `yaml:"templates"`
	} `yaml:"deep_links"`
	EventLinks map[string]string `yaml:"event_deep_links"`
	Worker     struct {
		Concurrency int           `yaml:"concurrency"`
		RetryMax    int           `yaml:"retry_max"`
		RetryDelay  time.Duration `yaml:"retry_delay"`
	} `yaml:"worker"`
	Batch struct {
		MaxSize       int           `yaml:"max_size"`
		FlushInterval time.Duration `yaml:"flush_interval"`
	} `yaml:"batch"`
	QuietHours struct {
		CheckEnabled    bool   `yaml:"check_enabled"`
		DefaultTimezone string `yaml:"default_timezone"`
	} `yaml:"quiet_hours"`
}
