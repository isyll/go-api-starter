package cache

import "fmt"

// Cache-key templates. Every cached value must be built through one
// of these helpers so the key vocabulary stays in one place.
const (
	// KeyVerificationToken stores email-verification tokens.
	KeyVerificationToken = "auth:verify:%s"
	// KeyPasswordReset stores password-reset tokens.
	KeyPasswordReset = "auth:reset:%s"
	// KeyRateLimit is a generic rate-limit slot.
	KeyRateLimit = "rate:%s:%s"
	// keySessionData caches a *models.DeviceSession by ID.
	keySessionData = "session:%d"
	// keyAdminInvalidated marks an admin's service tokens invalid
	// until natural expiry.
	keyAdminInvalidated = "admin:invalidated:%s"
)

// SessionDataKey returns the cache key for a device session.
func SessionDataKey(sessionID int64) string {
	return fmt.Sprintf(keySessionData, sessionID)
}

// VerificationTokenKey returns the key for an email-verification token.
func VerificationTokenKey(token string) string {
	return fmt.Sprintf(KeyVerificationToken, token)
}

// PasswordResetKey returns the key for a password-reset token.
func PasswordResetKey(token string) string {
	return fmt.Sprintf(KeyPasswordReset, token)
}

// RateLimitKey returns a generic rate-limit slot key.
func RateLimitKey(limitType, identifier string) string {
	return fmt.Sprintf(KeyRateLimit, limitType, identifier)
}

// AdminInvalidatedKey returns the key that forces an admin to
// re-authenticate when present.
func AdminInvalidatedKey(adminSub string) string {
	return fmt.Sprintf(keyAdminInvalidated, adminSub)
}
