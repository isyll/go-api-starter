package auth

import (
	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
)

// Auth domain sentinel errors.
var (
	ErrInvalidCredentials = apperrors.Unauthorized(
		codes.InvalidCredentials, "auth.invalid_credentials",
	)
	ErrEmailExists = apperrors.Conflict(
		codes.EmailAlreadyExists, "auth.email_exists",
	)
	ErrEmailNotVerified = apperrors.Forbidden(
		codes.EmailNotVerified, "auth.email_not_verified",
	)
	ErrAccountInactive = apperrors.Forbidden(
		codes.AccountInactive, "auth.account_inactive",
	)
	ErrInvalidToken = apperrors.Unauthorized(
		codes.InvalidAuthToken, "auth.invalid_token",
	)
	ErrTokenRevoked = apperrors.Unauthorized(
		codes.AccessTokenRevoked, "auth.token_revoked",
	)
	ErrSessionNotFound = apperrors.Unauthorized(
		codes.SessionNotFound, "auth.session_not_found",
	)
	ErrCannotRemoveCurrentDevice = apperrors.BadRequest(
		codes.CannotRemoveCurrentDevice, "auth.cannot_remove_current_device",
	)
	ErrInvalidVerificationToken = apperrors.BadRequest(
		codes.InvalidVerificationToken, "auth.invalid_verification_token",
	)
	ErrInvalidResetToken = apperrors.BadRequest(
		codes.InvalidResetToken, "auth.invalid_reset_token",
	)
	ErrIncorrectPassword = apperrors.BadRequest(
		codes.IncorrectPassword, "auth.incorrect_password",
	)
	ErrWeakPassword = apperrors.BadRequest(
		codes.WeakPassword, "auth.weak_password",
	)
)
