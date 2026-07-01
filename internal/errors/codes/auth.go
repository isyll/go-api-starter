package codes

// Auth-domain error codes.
var (
	// InvalidCredentials - 401. Email or password did not match.
	InvalidCredentials = register("INVALID_CREDENTIALS")

	// EmailAlreadyExists - 409. Registration email is already taken.
	EmailAlreadyExists = register("EMAIL_ALREADY_EXISTS")

	// EmailNotVerified - 403. Action requires a verified email.
	EmailNotVerified = register("EMAIL_NOT_VERIFIED")

	// AccountInactive - 403. User exists but status is not active.
	AccountInactive = register("ACCOUNT_INACTIVE")

	// InvalidAuthToken - 401. Access/refresh token missing or expired.
	InvalidAuthToken = register("INVALID_AUTH_TOKEN")

	// AccessTokenRevoked - 401. Reuse of a revoked refresh token.
	AccessTokenRevoked = register("ACCESS_TOKEN_REVOKED")

	// SessionNotFound - 401. Session not found (revoked/expired/unknown).
	SessionNotFound = register("SESSION_NOT_FOUND")

	// CannotRemoveCurrentDevice - 400. Use logout for the current session.
	CannotRemoveCurrentDevice = register("CANNOT_REMOVE_CURRENT_DEVICE")

	// InvalidVerificationToken - 400. Email-verification token invalid.
	InvalidVerificationToken = register("INVALID_VERIFICATION_TOKEN")

	// InvalidResetToken - 400. Password-reset token invalid or expired.
	InvalidResetToken = register("INVALID_RESET_TOKEN")

	// IncorrectPassword - 400. Current password did not match.
	IncorrectPassword = register("INCORRECT_PASSWORD")

	// WeakPassword - 400. Password does not meet strength requirements.
	WeakPassword = register("WEAK_PASSWORD")
)
