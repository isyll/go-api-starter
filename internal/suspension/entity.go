package suspension

import "time"

type SuspensionReason string

const (
	SuspensionReasonTermsViolation     SuspensionReason = "terms_violation"
	SuspensionReasonFraudulentActivity SuspensionReason = "fraudulent_activity"
	SuspensionReasonHarassment         SuspensionReason = "harassment"
	SuspensionReasonSpam               SuspensionReason = "spam"
	SuspensionReasonSecurityBreach     SuspensionReason = "security_breach"
	SuspensionReasonLegalRequest       SuspensionReason = "legal_request"
	SuspensionReasonOther              SuspensionReason = "other"
)

type AccountSuspension struct {
	ID             int64
	UserID         int64
	Reason         SuspensionReason
	Details        string
	SuspendedAt    time.Time
	SuspendedUntil *time.Time
	IsPermanent    bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *AccountSuspension) IsActive() bool {
	if s.IsPermanent {
		return true
	}
	return s.SuspendedUntil != nil && s.SuspendedUntil.After(time.Now())
}
