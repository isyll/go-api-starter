package config

import (
	"errors"
	"fmt"
	"strings"
)

type Configs struct {
	App           *AppConfig
	Database      *DatabaseConfig
	Redis         *RedisConfig
	Security      *SecurityConfig
	Notifications *NotificationsConfig
	Email         *EmailConfig
	Events        *EventsConfig
	Storage       *StorageConfig
	Firebase      *FirebaseConfig
	Gateway       *GatewayConfig
}

// LoadAllConfigs is the single entrypoint: it loads every config file from
// configs/, applies defaults, and validates the result. Every file is
// required; a missing or malformed file is a hard error.
func LoadAllConfigs() (*Configs, error) {
	LoadEnvFile()

	app, err := LoadConfig[AppConfig]("configs/api.yaml")
	if err != nil {
		return nil, fmt.Errorf("api config: %w", err)
	}
	db, err := LoadConfig[DatabaseConfig]("configs/database.yaml")
	if err != nil {
		return nil, fmt.Errorf("database config: %w", err)
	}
	redis, err := LoadConfig[RedisConfig]("configs/redis.yaml")
	if err != nil {
		return nil, fmt.Errorf("redis config: %w", err)
	}
	security, err := LoadConfig[SecurityConfig]("configs/security.yaml")
	if err != nil {
		return nil, fmt.Errorf("security config: %w", err)
	}
	notifications, err := LoadConfig[NotificationsConfig]("configs/notifications.yaml")
	if err != nil {
		return nil, fmt.Errorf("notifications config: %w", err)
	}
	email, err := LoadConfig[EmailConfig]("configs/email.yaml")
	if err != nil {
		return nil, fmt.Errorf("email config: %w", err)
	}
	events, err := LoadConfig[EventsConfig]("configs/events.yaml")
	if err != nil {
		return nil, fmt.Errorf("events config: %w", err)
	}
	events.applyDefaults()
	storage, err := LoadConfig[StorageConfig]("configs/storage.yaml")
	if err != nil {
		return nil, fmt.Errorf("storage config: %w", err)
	}
	firebase, err := LoadConfig[FirebaseConfig]("configs/firebase.yaml")
	if err != nil {
		return nil, fmt.Errorf("firebase config: %w", err)
	}
	gateway, err := LoadConfig[GatewayConfig]("configs/gateway.yaml")
	if err != nil {
		return nil, fmt.Errorf("gateway config: %w", err)
	}

	cfg := &Configs{
		App:           app,
		Database:      db,
		Redis:         redis,
		Security:      security,
		Notifications: notifications,
		Email:         email,
		Events:        events,
		Storage:       storage,
		Firebase:      firebase,
		Gateway:       gateway,
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

// Validate checks the invariants the process needs to start.
func (c *Configs) Validate() error {
	var problems []string
	if c.App.Server.Port == "" {
		problems = append(problems, "server port is required")
	}
	if c.App.Info.Name == "" {
		problems = append(problems, "app name is required")
	}
	if c.Database.AppCredentials.Host == "" || c.Database.AppCredentials.DBName == "" {
		problems = append(problems, "database app credentials (host, dbname) are required")
	}
	if c.Redis.Connection.Host == "" {
		problems = append(problems, "redis host is required")
	}
	if c.Security.IDObfuscation.Alphabet == "" || c.Security.IDObfuscation.MinLength <= 0 {
		problems = append(problems, "id_obfuscation (alphabet, min_length) is required")
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}
