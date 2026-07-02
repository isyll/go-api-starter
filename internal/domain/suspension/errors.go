package suspension

import (
	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
)

var ErrNotSuspended = apperrors.NotFound(
	codes.SuspensionNotFound,
	"user.suspension.not_suspended",
)
