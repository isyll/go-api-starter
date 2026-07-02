package models

import "time"

type LoginAttempt struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email"`
	UserID    int64     `json:"user_id,omitempty"`
	Channel   string    `json:"channel"`
	Outcome   string    `json:"outcome"`
	Remaining int       `json:"remaining,omitempty"`
	IPAddress *string   `json:"ip_address,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
	DeviceID  *string   `json:"device_id,omitempty"`
	RequestID *string   `json:"request_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (LoginAttempt) TableName() string {
	return "auth.login_attempts"
}
