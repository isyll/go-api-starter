package auth

import (
	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/errs/codes"
)

var (
	ErrInvalidCredentials = errs.Unauthorized(
		codes.InvalidCredentials, "auth.invalid_credentials",
	)
	ErrEmailExists = errs.Conflict(
		codes.EmailAlreadyExists, "auth.email_exists",
	)
	ErrEmailNotVerified = errs.Forbidden(
		codes.EmailNotVerified, "auth.email_not_verified",
	)
	ErrAccountInactive = errs.Forbidden(
		codes.AccountInactive, "auth.account_inactive",
	)
	ErrInvalidToken = errs.Unauthorized(
		codes.InvalidAuthToken, "auth.invalid_token",
	)
	ErrTokenRevoked = errs.Unauthorized(
		codes.AccessTokenRevoked, "auth.token_revoked",
	)
	ErrSessionNotFound = errs.Unauthorized(
		codes.SessionNotFound, "auth.session_not_found",
	)
	ErrCannotRemoveCurrentDevice = errs.BadRequest(
		codes.CannotRemoveCurrentDevice, "auth.cannot_remove_current_device",
	)
	ErrInvalidVerificationToken = errs.BadRequest(
		codes.InvalidVerificationToken, "auth.invalid_verification_token",
	)
	ErrInvalidResetToken = errs.BadRequest(
		codes.InvalidResetToken, "auth.invalid_reset_token",
	)
	ErrIncorrectPassword = errs.BadRequest(
		codes.IncorrectPassword, "auth.incorrect_password",
	)
	ErrWeakPassword = errs.BadRequest(
		codes.WeakPassword, "auth.weak_password",
	)
)
