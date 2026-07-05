package cache

import "fmt"

const (
	KeyVerificationToken = "auth:verify:%s" //nolint:gosec // cache key format, not a credential
	KeyPasswordReset     = "auth:reset:%s"  //nolint:gosec // cache key format, not a credential
	KeyRateLimit         = "rate:%s:%s"
	keySessionData       = "session:%d"
	keyAdminInvalidated  = "admin:invalidated:%s"
)

func SessionDataKey(sessionID int64) string {
	return fmt.Sprintf(keySessionData, sessionID)
}

func VerificationTokenKey(token string) string {
	return fmt.Sprintf(KeyVerificationToken, token)
}

func PasswordResetKey(token string) string {
	return fmt.Sprintf(KeyPasswordReset, token)
}

func RateLimitKey(limitType, identifier string) string {
	return fmt.Sprintf(KeyRateLimit, limitType, identifier)
}

func AdminInvalidatedKey(adminSub string) string {
	return fmt.Sprintf(keyAdminInvalidated, adminSub)
}
