package users

import (
	apperrors "github.com/isyll/go-grpc-starter/internal/errors"
	"github.com/isyll/go-grpc-starter/internal/errors/codes"
)

var (
	ErrUserNotFound  = apperrors.ErrUserNotFound
	ErrInvalidUserID = apperrors.BadRequest(codes.InvalidUserID, "user.invalid_id")
)
