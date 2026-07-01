package apperrors

import "github.com/isyll/go-api-starter/internal/errors/codes"

// Cross-domain sentinel errors shared by multiple packages.
// Domain-specific sentinels live in each domain's errors.go.
var (
	// ErrInvalidID is returned when a path/argument does not decode
	// to a valid internal ID.
	ErrInvalidID = BadRequest(codes.InvalidID, "common.invalid_id")

	// ErrForbidden is the generic 403.
	ErrForbidden = Forbidden(codes.Forbidden, "common.forbidden")

	// ErrNoFieldsToUpdate is a 400 returned when an update carries no
	// mutable fields.
	ErrNoFieldsToUpdate = BadRequest(
		codes.NoFieldsToUpdate, "common.no_fields_to_update",
	)

	// ErrInvalidDateRange is a 400 for a reversed or inconsistent
	// date range.
	ErrInvalidDateRange = BadRequest(
		codes.InvalidDateRange, "common.date_invalid_range",
	)

	// ErrStartDateInPast is a 400 when a future-only date is in the
	// past.
	ErrStartDateInPast = BadRequest(codes.DateInPast, "common.date_in_past")

	// ErrSessionNotFound is a 401 when the session behind the access
	// token can no longer be located.
	ErrSessionNotFound = Unauthorized(
		codes.SessionNotFound, "auth.session_not_found",
	)

	// ErrUserNotFound is a 404 when a user lookup finds no visible row.
	ErrUserNotFound = NotFound(codes.UserNotFound, "user.not_found")
)
