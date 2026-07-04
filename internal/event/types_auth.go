package event

import "time"

type AuthAttemptRecorded struct {
	Email      string    `json:"email"`
	UserID     int64     `json:"user_id,omitempty"`
	Channel    string    `json:"channel"`
	Outcome    string    `json:"outcome"`
	Remaining  int       `json:"remaining,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	DeviceID   string    `json:"device_id,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (*AuthAttemptRecorded) EventType() string {
	return "auth.attempt.recorded"
}

func (e *AuthAttemptRecorded) OutboxDedupeKey() string {
	if e.RequestID == "" {
		return ""
	}
	return e.RequestID + ":" + e.Channel + ":" + e.Outcome
}
