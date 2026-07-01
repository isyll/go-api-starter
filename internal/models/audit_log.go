package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// AuditDetails is a JSONB map for free-form metadata on an audit entry.
type AuditDetails map[string]any

func (d AuditDetails) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

func (d *AuditDetails) Scan(value any) error {
	if value == nil {
		*d = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}
	return json.Unmarshal(b, d)
}

// AuditLog is the GORM model for audit.audit_logs. Every admin action
// is appended here; rows are never updated or deleted.
type AuditLog struct {
	ID         int64        `gorm:"primaryKey"               json:"id"`
	AdminID    int64        `gorm:"not null"                 json:"admin_id"`
	Action     string       `gorm:"not null"                 json:"action"`
	Resource   string       `gorm:"not null"                 json:"resource"`
	ResourceID string       `                                json:"resource_id,omitempty"`
	Details    AuditDetails `gorm:"type:jsonb"               json:"details,omitempty"`
	Status     string       `gorm:"not null;default:success" json:"status"`
	IPAddress  string       `                                json:"ip_address,omitempty"`
	UserAgent  string       `                                json:"user_agent,omitempty"`
	RequestID  string       `                                json:"request_id,omitempty"`
	CreatedAt  time.Time    `gorm:"autoCreateTime"           json:"created_at"`
}

func (AuditLog) TableName() string { return "audit.audit_logs" }
