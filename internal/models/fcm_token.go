package models

import "time"

// NotificationPlatform identifies the OS targeted by a push token.
type NotificationPlatform string // @name NotificationPlatform

const (
	PlatformAndroid NotificationPlatform = "android"
	PlatformIOS     NotificationPlatform = "ios"
	PlatformWeb     NotificationPlatform = "web"
)

// FCMToken is the GORM model for auth.fcm_tokens. Each FCM token
// targets one device. Tokens are rotated by the client on refresh;
// stale tokens are deactivated (IsActive = false) when FCM reports
// a delivery failure.
type FCMToken struct {
	ID         int64                `gorm:"primaryKey" json:"id"                     msgpack:"id"`
	UserID     int64                `                  json:"user_id"                msgpack:"user_id"`
	DeviceID   string               `                  json:"device_id"              msgpack:"device_id"`
	Token      string               `                  json:"token"                  msgpack:"token"`
	Platform   NotificationPlatform `                  json:"platform"               msgpack:"platform"`
	AppVersion string               `                  json:"app_version,omitempty"  msgpack:"app_version,omitempty"`
	IsActive   bool                 `                  json:"is_active"              msgpack:"is_active"`
	LastUsedAt *time.Time           `                  json:"last_used_at,omitempty" msgpack:"last_used_at,omitempty"`
	CreatedAt  time.Time            `                  json:"created_at"             msgpack:"created_at"`
	UpdatedAt  time.Time            `                  json:"updated_at"             msgpack:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty" msgpack:"user,omitempty"`
}

// TableName returns the schema-qualified table name for GORM.
func (FCMToken) TableName() string {
	return "auth.fcm_tokens"
}
