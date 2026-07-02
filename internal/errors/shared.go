package apperrors

import "github.com/isyll/go-api-starter/internal/errors/codes"

var (
	ErrInvalidID = BadRequest(codes.InvalidID, "common.invalid_id")

	ErrForbidden = Forbidden(codes.Forbidden, "common.forbidden")

	ErrNoFieldsToUpdate = BadRequest(
		codes.NoFieldsToUpdate, "common.no_fields_to_update",
	)

	ErrInvalidDateRange = BadRequest(
		codes.InvalidDateRange, "common.date_invalid_range",
	)

	ErrStartDateInPast = BadRequest(codes.DateInPast, "common.date_in_past")

	ErrSessionNotFound = Unauthorized(
		codes.SessionNotFound, "auth.session_not_found",
	)

	ErrUserNotFound = NotFound(codes.UserNotFound, "user.not_found")
)
