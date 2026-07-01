package suspension

import (
	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
)

// ErrNotSuspended — 404. Returned by the suspension helper when
// no active suspension exists for the user. Code pinned to
// internal/errors/codes.
var ErrNotSuspended = apperrors.NotFound(
	codes.SuspensionNotFound,
	"user.suspension.not_suspended",
)
