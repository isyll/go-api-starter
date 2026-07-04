package event

import "time"

type AuditLogWritten struct {
	AdminID    int64          `json:"admin_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resource_id,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	Status     string         `json:"status"`
	IPAddress  string         `json:"ip_address,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	OccurredAt time.Time      `json:"occurred_at"`
}

func (*AuditLogWritten) EventType() string {
	return "admin.audit_log.written"
}

func (e *AuditLogWritten) OutboxDedupeKey() string {
	if e.RequestID == "" {
		return ""
	}
	return e.RequestID + ":" + e.Action + ":" + e.ResourceID
}
