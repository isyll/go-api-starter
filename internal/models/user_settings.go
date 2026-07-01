package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Theme is the user's preferred UI theme.
type Theme string // @name Theme

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

// IsValid reports whether t is a known Theme value.
func (t Theme) IsValid() bool {
	switch t {
	case ThemeSystem, ThemeLight, ThemeDark:
		return true
	}
	return false
}

// Settings is the JSONB value stored in UserSettings.Settings.
type Settings struct {
	Locale             string `json:"locale"              msgpack:"locale"`
	Timezone           string `json:"timezone"            msgpack:"timezone"`
	Theme              Theme  `json:"theme"               msgpack:"theme"`
	EmailNotifications bool   `json:"email_notifications" msgpack:"email_notifications"`
	PushNotifications  bool   `json:"push_notifications"  msgpack:"push_notifications"`
	MarketingEmails    bool   `json:"marketing_emails"    msgpack:"marketing_emails"`
} // @name Settings

// DefaultSettings returns the settings applied to a new user.
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

// UserSettings is the GORM model for auth.user_settings. One row per
// user, created at registration.
type UserSettings struct {
	UserID    int64     `json:"user_id"    msgpack:"user_id"    gorm:"primaryKey"`
	Settings  Settings  `json:"settings"   msgpack:"settings"`
	CreatedAt time.Time `json:"created_at" msgpack:"created_at"`
	UpdatedAt time.Time `json:"updated_at" msgpack:"updated_at"`
} // @name UserSettings

func (UserSettings) TableName() string {
	return "auth.user_settings"
}
