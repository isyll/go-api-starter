package users

import (
	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/errs/codes"
)

var (
	ErrUserNotFound  = errs.ErrUserNotFound
	ErrInvalidUserID = errs.BadRequest(codes.InvalidUserID, "user.invalid_id")
)
