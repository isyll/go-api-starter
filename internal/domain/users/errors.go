package users

import (
	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
)

// Sentinel errors returned by the users domain.
var (
	// ErrUserNotFound - 404. Row absent or hidden by RLS.
	ErrUserNotFound = apperrors.ErrUserNotFound
	// ErrInvalidUserID - 400. Encoded user id failed to decode.
	ErrInvalidUserID = apperrors.BadRequest(codes.InvalidUserID, "user.invalid_id")
)
