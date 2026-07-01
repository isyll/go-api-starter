package models

import "time"

// LoginAttempt is the GORM model for auth.login_attempts. Every auth
// step (login, token refresh, password reset) appends a row here
// regardless of outcome. The table is append-only at the DB layer.
type LoginAttempt struct {
	ID int64 `gorm:"primaryKey" json:"id"`
	// Email is the target account email (lowercased).
	Email string `json:"email"`
	// UserID is the resolved account ID, or zero when unknown.
	UserID int64 `json:"user_id,omitempty"`
	// Channel is the auth step: login, refresh, password_reset.
	Channel string `json:"channel"`
	// Outcome is the result: success, wrong_password, not_found,
	// rate_limited, blocked.
	Outcome string `json:"outcome"`
	// Remaining is the attempts left after a wrong_password outcome.
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
