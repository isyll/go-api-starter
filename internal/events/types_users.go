package events

import "time"

// UserAccountDeleted is published when a user soft-deletes their own
// account, after the account teardown (status flip, email purge, and
// the cascade cancellation of the user's active trips and bookings)
// has committed. It exists to drive cross-cutting reactions that are
// NOT already covered by the trip/booking cancellation events — chiefly
// cache invalidation of the now-gone user's own cached surface.
//
// The cancellation of the user's trips and bookings is performed
// synchronously in the deletion request (reusing the trip/booking
// cancellation flows, which publish their own TripCancelled /
// BookingStatusChanged events for notifications). This event therefore
// does NOT itself fan out cancellations — doing so from an async
// handler is impossible because the event-dispatcher worker's bus has
// no async dispatcher and may not re-publish.
type UserAccountDeleted struct {
	// UserID is the internal id of the deleted account.
	UserID int64 `json:"user_id"`
	// AccountID is the public account number, retained for audit
	// correlation in downstream consumers.
	AccountID int64 `json:"account_id"`
	// RequestID correlates async reactions with the originating
	// request's log lines.
	RequestID string `json:"request_id,omitempty"`
	// OccurredAt is the deletion commit time (UTC).
	OccurredAt time.Time `json:"occurred_at"`
}

// EventType returns the bus routing key "user.account.deleted".
func (*UserAccountDeleted) EventType() string {
	return "user.account.deleted"
}
