package config

import "fmt"

// Configs aggregates every subsystem configuration loaded at startup
// by LoadAllConfigs.
type Configs struct {
	App           *AppConfig
	Database      *DatabaseConfig
	Redis         *RedisConfig
	Security      *SecurityConfig
	Notifications *NotificationsConfig
	Email         *EmailConfig
	Events        *EventsConfig
}

// LoadAllConfigs reads every YAML file under configs/, substitutes
// ${ENV_VAR:-default} placeholders, and returns a populated *Configs.
// events.yaml is optional and falls back to EventsDefaults.
func LoadAllConfigs() (*Configs, error) {
	LoadEnvFile()

	app, err := LoadConfig[AppConfig]("configs/api.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load api config: %w", err)
	}

	db, err := LoadConfig[DatabaseConfig]("configs/database.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load db config: %w", err)
	}

	redis, err := LoadConfig[RedisConfig]("configs/redis.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load redis config: %w", err)
	}

	security, err := LoadConfig[SecurityConfig]("configs/security.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load security config: %w", err)
	}

	notifications, err := LoadConfig[NotificationsConfig](
		"configs/notifications.yaml",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load notifications config: %w", err)
	}

	email, err := LoadConfig[EmailConfig]("configs/email.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load email config: %w", err)
	}

	// events.yaml is optional; fall back to defaults and treat
	// zero-valued fields as "use the default".
	eventsCfg, eventsErr := LoadConfig[EventsConfig]("configs/events.yaml")
	if eventsErr != nil {
		eventsCfg = EventsDefaults()
	} else {
		defaults := EventsDefaults()
		if eventsCfg.Outbox.Interval == 0 {
			eventsCfg.Outbox.Interval = defaults.Outbox.Interval
		}
		if eventsCfg.Outbox.MetricsInterval == 0 {
			eventsCfg.Outbox.MetricsInterval = defaults.Outbox.MetricsInterval
		}
	}

	return &Configs{
		App:           app,
		Database:      db,
		Redis:         redis,
		Security:      security,
		Notifications: notifications,
		Email:         email,
		Events:        eventsCfg,
	}, nil
}
