package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NotificationStatus tracks the delivery and interaction lifecycle
// of a push notification.
type NotificationStatus string // @name NotificationStatus

const (
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusClicked   NotificationStatus = "clicked"
	NotificationStatusDismissed NotificationStatus = "dismissed"
)

// NotificationLog is the GORM model for
// notifications.notification_logs. One row per FCM send attempt.
// EventID links the notification back to the originating domain
// event for cross-system correlation.
type NotificationLog struct {
	ID           int64              `gorm:"primaryKey" json:"id"                      msgpack:"id"`
	UserID       *int64             `                  json:"user_id,omitempty"       msgpack:"user_id,omitempty"`
	EventType    string             `                  json:"event_type"              msgpack:"event_type"`
	EventID      *string            `                  json:"event_id,omitempty"      msgpack:"event_id,omitempty"`
	FCMTokenID   *int64             `                  json:"fcm_token_id,omitempty"  msgpack:"fcm_token_id,omitempty"`
	Status       NotificationStatus `                  json:"status"                  msgpack:"status"`
	ErrorCode    *string            `                  json:"error_code,omitempty"    msgpack:"error_code,omitempty"`
	ErrorMessage *string            `                  json:"error_message,omitempty" msgpack:"error_message,omitempty"`
	Payload      *JSONB             `                  json:"payload,omitempty"       msgpack:"payload,omitempty"`
	SentAt       time.Time          `                  json:"sent_at"                 msgpack:"sent_at"`
	ClickedAt    *time.Time         `                  json:"clicked_at,omitempty"    msgpack:"clicked_at,omitempty"`
	DismissedAt  *time.Time         `                  json:"dismissed_at,omitempty"  msgpack:"dismissed_at,omitempty"`

	User     *User     `gorm:"foreignKey:UserID;references:ID"     json:"user,omitempty"      msgpack:"user,omitempty"`
	FCMToken *FCMToken `gorm:"foreignKey:FCMTokenID;references:ID" json:"fcm_token,omitempty" msgpack:"fcm_token,omitempty"`
}

// TableName returns the schema-qualified table name for GORM.
func (NotificationLog) TableName() string {
	return "notifications.notification_logs"
}

// JSONB is a PostgreSQL JSONB column type that marshals/unmarshals
// to and from a map[string]any.
type JSONB map[string]any

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
