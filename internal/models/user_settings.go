package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Theme string

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

func (t Theme) IsValid() bool {
	switch t {
	case ThemeSystem, ThemeLight, ThemeDark:
		return true
	}
	return false
}

type Settings struct {
	Locale             string `json:"locale"              msgpack:"locale"`
	Timezone           string `json:"timezone"            msgpack:"timezone"`
	Theme              Theme  `json:"theme"               msgpack:"theme"`
	EmailNotifications bool   `json:"email_notifications" msgpack:"email_notifications"`
	PushNotifications  bool   `json:"push_notifications"  msgpack:"push_notifications"`
	MarketingEmails    bool   `json:"marketing_emails"    msgpack:"marketing_emails"`
}

func DefaultSettings() Settings {
	return Settings{
		Locale:             "en",
		Timezone:           "UTC",
		Theme:              ThemeSystem,
		EmailNotifications: true,
		PushNotifications:  true,
		MarketingEmails:    false,
	}
}

func (s *Settings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Settings) Scan(value any) error {
	if value == nil {
		*s = DefaultSettings()
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Settings: invalid type")
	}
	return json.Unmarshal(bytes, s)
}

type UserSettings struct {
	UserID    int64     `json:"user_id"    msgpack:"user_id"    gorm:"primaryKey"`
	Settings  Settings  `json:"settings"   msgpack:"settings"`
	CreatedAt time.Time `json:"created_at" msgpack:"created_at"`
	UpdatedAt time.Time `json:"updated_at" msgpack:"updated_at"`
}

func (UserSettings) TableName() string {
	return "auth.user_settings"
}
