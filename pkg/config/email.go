package config

import "time"

type EmailConfig struct {
	Email struct {
		Provider string `yaml:"provider"`
		APIKey   string `yaml:"api_key"`
		Senders  struct {
			NoReply  SenderInfo `yaml:"noreply"`
			Security SenderInfo `yaml:"security"`
			Support  SenderInfo `yaml:"support"`
			Billing  SenderInfo `yaml:"billing"`
			News     SenderInfo `yaml:"news"`
		} `yaml:"senders"`
		Worker struct {
			Concurrency int           `yaml:"concurrency"`
			RetryMax    int           `yaml:"retry_max"`
			RetryDelay  time.Duration `yaml:"retry_delay"`
		} `yaml:"worker"`
		RateLimit struct {
			PerSecond int `yaml:"per_second"`
			Burst     int `yaml:"burst"`
		} `yaml:"rate_limit"`
		Templates struct {
			BasePath        string `yaml:"base_path"`
			DefaultLanguage string `yaml:"default_language"`
		} `yaml:"templates"`
		Batch struct {
			MaxSize       int           `yaml:"max_size"`
			FlushInterval time.Duration `yaml:"flush_interval"`
		} `yaml:"batch"`
	} `yaml:"email"`
}

type SenderInfo struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}
