package suspension

import (
	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/errs/codes"
)

var ErrNotSuspended = errs.NotFound(
	codes.SuspensionNotFound,
	"user.suspension.not_suspended",
)
