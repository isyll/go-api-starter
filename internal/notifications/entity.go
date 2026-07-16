package notifications

import "time"

type NotificationPlatform string

const (
	PlatformAndroid NotificationPlatform = "android"
	PlatformIOS     NotificationPlatform = "ios"
	PlatformWeb     NotificationPlatform = "web"
)

type NotificationStatus string

const (
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusClicked   NotificationStatus = "clicked"
	NotificationStatusDismissed NotificationStatus = "dismissed"
)

type JSONB map[string]any

type FCMToken struct {
	ID         int64
	UserID     int64
	DeviceID   string
	Token      string
	Platform   NotificationPlatform
	AppVersion string
	IsActive   bool
	LastUsedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NotificationPreferences struct {
	UserID            int64
	Push              bool
	Email             bool
	Marketing         bool
	QuietHoursEnabled bool
	QuietHoursStart   *string
	QuietHoursEnd     *string
	Timezone          string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type NotificationTemplate struct {
	ID               int
	EventType        string
	Icon             *string
	Sound            string
	Priority         string
	AndroidChannelID *string
	Action           *string
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Translations []*NotificationTemplateTranslation
}

type NotificationTemplateTranslation struct {
	ID         int
	TemplateID int
	Language   string
	Title      string
	Body       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NotificationLog struct {
	ID           int64
	UserID       *int64
	EventType    string
	EventID      *string
	FCMTokenID   *int64
	Status       NotificationStatus
	ErrorCode    *string
	ErrorMessage *string
	Payload      *JSONB
	SentAt       time.Time
	ClickedAt    *time.Time
	DismissedAt  *time.Time
}
