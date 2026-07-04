package settings

import "time"

type Theme string

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

// Settings is the JSON preference bag stored in auth.user_settings.settings.
type Settings struct {
	Locale             string `json:"locale"`
	Timezone           string `json:"timezone"`
	Theme              Theme  `json:"theme"`
	EmailNotifications bool   `json:"email_notifications"`
	PushNotifications  bool   `json:"push_notifications"`
	MarketingEmails    bool   `json:"marketing_emails"`
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

// UserSettings is the settings row keyed by user id.
type UserSettings struct {
	UserID    int64
	Settings  Settings
	CreatedAt time.Time
	UpdatedAt time.Time
}
