package notifications

import (
	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
)

var (
	ErrTokenNotFound = apperrors.NotFound(
		codes.PushTokenNotFound,
		"notification.token_not_found",
	)
	ErrTokenInactive = apperrors.BadRequest(
		codes.PushTokenInactive,
		"notification.token_inactive",
	)
	ErrDeviceMismatch = apperrors.Forbidden(
		codes.DeviceMismatch,
		"notification.device_mismatch",
	)
	ErrPrefNotFound = apperrors.NotFound(
		codes.NotificationPrefsNotFound,
		"notification.prefs_not_found",
	)
	ErrInvalidTimezone = apperrors.BadRequest(
		codes.InvalidTimezone,
		"notification.invalid_timezone",
	)
	ErrSendFailed = apperrors.Internal(
		codes.PushSendFailed,
		"notification.send_failed",
	)
)
