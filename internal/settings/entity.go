package settings

import "time"

type Theme string

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

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

type UserSettings struct {
	UserID    int64
	Settings  Settings
	CreatedAt time.Time
	UpdatedAt time.Time
}
