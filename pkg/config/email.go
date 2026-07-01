package config

import "time"

// EmailConfig holds Resend API credentials, sender identities,
// worker concurrency, rate-limit, and email template settings
// loaded from configs/email.yaml.
type EmailConfig struct {
	Email struct {
		// Provider names the transactional email vendor (e.g. "resend").
		Provider string `yaml:"provider"`
		// APIKey is the vendor API key used for authenticated delivery.
		APIKey string `yaml:"api_key"`
		// Senders lists the named sender identities available to the
		// email worker. Each role has a distinct From address.
		Senders struct {
			NoReply  SenderInfo `yaml:"noreply"`
			Security SenderInfo `yaml:"security"`
			Support  SenderInfo `yaml:"support"`
			Billing  SenderInfo `yaml:"billing"`
			News     SenderInfo `yaml:"news"`
		} `yaml:"senders"`
		// Worker holds concurrency and retry settings for the Asynq
		// email processor.
		Worker struct {
			Concurrency int           `yaml:"concurrency"`
			RetryMax    int           `yaml:"retry_max"`
			RetryDelay  time.Duration `yaml:"retry_delay"`
		} `yaml:"worker"`
		// RateLimit caps the volume of outbound emails sent to the
		// vendor per second.
		RateLimit struct {
			PerSecond int `yaml:"per_second"`
			Burst     int `yaml:"burst"`
		} `yaml:"rate_limit"`
		// Templates describes where compiled HTML templates are stored
		// and which language to use when a translation is absent.
		Templates struct {
			BasePath        string `yaml:"base_path"`
			DefaultLanguage string `yaml:"default_language"`
		} `yaml:"templates"`
		// Batch controls how outbound emails are grouped before
		// flushing to the vendor API.
		Batch struct {
			MaxSize       int           `yaml:"max_size"`
			FlushInterval time.Duration `yaml:"flush_interval"`
		} `yaml:"batch"`
	} `yaml:"email"`
}

// SenderInfo holds a single named email sender identity composed of
// an RFC 5321 address and a human-readable display name.
type SenderInfo struct {
	// Address is the From email address in RFC 5321 format.
	Address string `yaml:"address"`
	// Name is the display name shown in email clients.
	Name string `yaml:"name"`
}
