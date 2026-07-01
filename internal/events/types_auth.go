package events

import "time"

// AuthAttemptRecorded fires after every authentication step (login,
// token refresh, password reset) regardless of outcome. The async
// handler appends a row to auth.login_attempts.
type AuthAttemptRecorded struct {
	// Email is the target account email (lowercased).
	Email string `json:"email"`
	// UserID is the resolved account ID, or zero if unknown.
	UserID int64 `json:"user_id,omitempty"`
	// Channel is the auth step: login, refresh, password_reset.
	Channel string `json:"channel"`
	// Outcome is the result: success, wrong_password, not_found,
	// rate_limited, blocked.
	Outcome string `json:"outcome"`
	// Remaining is the attempts left after a wrong_password outcome.
	Remaining  int       `json:"remaining,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	DeviceID   string    `json:"device_id,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventType returns the bus routing key.
func (*AuthAttemptRecorded) EventType() string {
	return "auth.attempt.recorded"
}

// OutboxDedupeKey collapses double-published attempts sharing the same
// request into one pending row. Empty RequestID opts out of dedupe.
func (e *AuthAttemptRecorded) OutboxDedupeKey() string {
	if e.RequestID == "" {
		return ""
	}
	return e.RequestID + ":" + e.Channel + ":" + e.Outcome
}
