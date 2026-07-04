package notifications

import (
	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/errs/codes"
)

var (
	ErrTokenNotFound = errs.NotFound(
		codes.PushTokenNotFound,
		"notification.token_not_found",
	)
	ErrTokenInactive = errs.BadRequest(
		codes.PushTokenInactive,
		"notification.token_inactive",
	)
	ErrDeviceMismatch = errs.Forbidden(
		codes.DeviceMismatch,
		"notification.device_mismatch",
	)
	ErrPrefNotFound = errs.NotFound(
		codes.NotificationPrefsNotFound,
		"notification.prefs_not_found",
	)
	ErrInvalidTimezone = errs.BadRequest(
		codes.InvalidTimezone,
		"notification.invalid_timezone",
	)
	ErrSendFailed = errs.Internal(
		codes.PushSendFailed,
		"notification.send_failed",
	)
)
