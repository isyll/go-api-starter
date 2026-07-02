package codes

var (
	InvalidCredentials = register("INVALID_CREDENTIALS")

	EmailAlreadyExists = register("EMAIL_ALREADY_EXISTS")

	EmailNotVerified = register("EMAIL_NOT_VERIFIED")

	AccountInactive = register("ACCOUNT_INACTIVE")

	InvalidAuthToken = register("INVALID_AUTH_TOKEN")

	AccessTokenRevoked = register("ACCESS_TOKEN_REVOKED")

	SessionNotFound = register("SESSION_NOT_FOUND")

	CannotRemoveCurrentDevice = register("CANNOT_REMOVE_CURRENT_DEVICE")

	InvalidVerificationToken = register("INVALID_VERIFICATION_TOKEN")

	InvalidResetToken = register("INVALID_RESET_TOKEN")

	IncorrectPassword = register("INCORRECT_PASSWORD")

	WeakPassword = register("WEAK_PASSWORD")
)
