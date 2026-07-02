package events

import "time"

type CommonFields struct {
	OccurredAt time.Time `json:"occurred_at"`
}

//nolint:gochecknoinits // factories must register at import time
func init() {
	Register(func() Event { return &UserAccountDeleted{} })
	Register(func() Event { return &AuditLogWritten{} })
	Register(func() Event { return &AuthAttemptRecorded{} })
}
