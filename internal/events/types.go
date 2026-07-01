package events

import "time"

// Concrete event types live in this package, split by
// origin domain into types_{domain}.go. Each type:
//
//   - has a pointer-receiver EventType() returning a
//     stable snake_case routing key,
//   - registers a Factory in init() so the async
//     processor can JSON-decode envelopes.
//
// Naming: domain.entity.action, past tense.
//
// Pointer receivers are mandatory because async decode
// allocates via the registry and unmarshals JSON into
// the returned value.

// CommonFields is embedded by every event so handlers
// always have a publication timestamp and a source
// user (best-effort) without redeclaring them per type.
type CommonFields struct {
	// OccurredAt is the wall-clock time the originating
	// service captured at publish, in UTC.
	OccurredAt time.Time `json:"occurred_at"`
}

// init populates the event-factory registry the async
// processor uses to JSON-decode payloads back into typed
// events. Registration MUST run at package import time so
// the registry is populated before any publisher or
// processor runs.
//
//nolint:gochecknoinits // factories must register at import time
func init() {
	Register(func() Event { return &UserAccountDeleted{} })
	Register(func() Event { return &AuditLogWritten{} })
	Register(func() Event { return &AuthAttemptRecorded{} })
}
