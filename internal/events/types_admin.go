package events

import "time"

// AuditLogWritten is published when a state-changing admin action
// completes. The async handler persists it to audit.audit_logs.
// Publishing through the bus gives the write outbox durability and
// keeps it off the request path.
type AuditLogWritten struct {
	// AdminID is the id of the admin who acted.
	AdminID int64 `json:"admin_id"`
	// Action is the verb, e.g. "user.suspend".
	Action string `json:"action"`
	// Resource is the resource type, e.g. "user".
	Resource string `json:"resource"`
	// ResourceID is the id of the affected resource, if any.
	ResourceID string `json:"resource_id,omitempty"`
	// Details holds before/after values or a reason.
	Details map[string]any `json:"details,omitempty"`
	// Status is the outcome: success, failed.
	Status     string    `json:"status"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventType returns the bus routing key.
func (*AuditLogWritten) EventType() string {
	return "admin.audit_log.written"
}

// OutboxDedupeKey collapses retries of the same admin action into one
// pending row. Empty RequestID opts out of dedupe.
func (e *AuditLogWritten) OutboxDedupeKey() string {
	if e.RequestID == "" {
		return ""
	}
	return e.RequestID + ":" + e.Action + ":" + e.ResourceID
}
