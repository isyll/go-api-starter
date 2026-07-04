package event

import "time"

type UserAccountDeleted struct {
	UserID     int64     `json:"user_id"`
	AccountID  int64     `json:"account_id"`
	RequestID  string    `json:"request_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (*UserAccountDeleted) EventType() string {
	return "user.account.deleted"
}
