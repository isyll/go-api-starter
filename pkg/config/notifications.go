package config

type NotificationsConfig struct {
	DeepLinks struct {
		BaseURL   string            `yaml:"base_url"`
		Templates map[string]string `yaml:"templates"`
	} `yaml:"deep_links"`
	EventLinks map[string]string `yaml:"event_deep_links"`
	Worker     struct {
		Concurrency int `yaml:"concurrency"`
		RetryMax    int `yaml:"retry_max"`
	} `yaml:"worker"`
	QuietHours struct {
		CheckEnabled    bool   `yaml:"check_enabled"`
		DefaultTimezone string `yaml:"default_timezone"`
	} `yaml:"quiet_hours"`
}
