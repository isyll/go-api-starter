package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/isyll/go-api-starter/pkg/idenc"
)

// SuspensionReason is the typed cause of an account suspension,
// recorded by the admin who imposed it.
type SuspensionReason string // @name SuspensionReason

const (
	SuspensionReasonTermsViolation     SuspensionReason = "terms_violation"
	SuspensionReasonFraudulentActivity SuspensionReason = "fraudulent_activity"
	SuspensionReasonHarassment         SuspensionReason = "harassment"
	SuspensionReasonSpam               SuspensionReason = "spam"
	SuspensionReasonSecurityBreach     SuspensionReason = "security_breach"
	SuspensionReasonLegalRequest       SuspensionReason = "legal_request"
	SuspensionReasonOther              SuspensionReason = "other"
)

// AccountSuspension is the GORM model for auth.account_suspensions.
// One active row per suspended user; the middleware checks IsActive()
// and returns 403 for every write request while the suspension holds.
type AccountSuspension struct {
	ID          int64            `gorm:"primaryKey" json:"id"              msgpack:"id"`
	UserID      int64            `                  json:"user_id"         msgpack:"user_id"`
	Reason      SuspensionReason `                  json:"reason"          msgpack:"reason"`
	Details     string           `                  json:"details"         msgpack:"details"`
	SuspendedAt time.Time        `                  json:"suspended_at"    msgpack:"suspended_at"`
	// SuspendedUntil is the expiry timestamp for time-limited
	// suspensions. A nil value indicates a permanent suspension
	// (IsPermanent must also be true in that case).
	SuspendedUntil *time.Time `                  json:"suspended_until" msgpack:"suspended_until"`
	// IsPermanent is true when the suspension has no expiry. When
	// true, SuspendedUntil is nil and IsActive always returns true.
	IsPermanent bool      `                  json:"is_permanent"    msgpack:"is_permanent"`
	CreatedAt   time.Time `                  json:"created_at"      msgpack:"created_at"`
	UpdatedAt   time.Time `                  json:"updated_at"      msgpack:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty" msgpack:"user,omitempty"`
}

// TableName returns the schema-qualified table name for GORM.
func (AccountSuspension) TableName() string {
	return "auth.account_suspensions"
}

// IsActive reports whether this suspension is currently in effect.
// A permanent suspension always returns true. A time-limited
// suspension returns true while the current time is before
// SuspendedUntil.
func (s *AccountSuspension) IsActive() bool {
	if s.IsPermanent {
		return true
	}
	if s.SuspendedUntil != nil && s.SuspendedUntil.After(time.Now()) {
		return true
	}
	return false
}

// BeforeCreate defaults SuspendedAt to the current time when the
// caller did not set an explicit suspension timestamp.
func (s *AccountSuspension) BeforeCreate(tx *gorm.DB) error {
	if s.SuspendedAt.IsZero() {
		s.SuspendedAt = time.Now()
	}
	return nil
}

// ToResponse converts the suspension to its Sqid-encoded API
// representation. encoder must not be nil.
func (s *AccountSuspension) ToResponse(
	encoder idenc.IDEncoder,
) *AccountSuspensionResponse {
	return &AccountSuspensionResponse{
		ID:             encoder.Encode(s.ID),
		Reason:         s.Reason,
		Details:        s.Details,
		SuspendedAt:    s.SuspendedAt,
		SuspendedUntil: s.SuspendedUntil,
		IsPermanent:    s.IsPermanent,
	}
}

// AccountSuspensionResponse is the API representation of an account
// suspension. The ID is Sqid-encoded; raw int64 IDs are never exposed.
type AccountSuspensionResponse struct {
	ID             string           `json:"id"                        msgpack:"id"                        example:"18n7q8765"`
	Reason         SuspensionReason `json:"reason"                    msgpack:"reason"                    example:"terms_violation"`
	Details        string           `json:"details"                   msgpack:"details"                   example:"Multiple violations of community guidelines"`
	SuspendedAt    time.Time        `json:"suspended_at"              msgpack:"suspended_at"              example:"2023-01-01T10:00:00Z"`
	SuspendedUntil *time.Time       `json:"suspended_until,omitempty" msgpack:"suspended_until,omitempty" example:"2023-02-01T10:00:00Z"`
	IsPermanent    bool             `json:"is_permanent"              msgpack:"is_permanent"              example:"false"`
} // @name AccountSuspensionResponse
