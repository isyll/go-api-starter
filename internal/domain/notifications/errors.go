package notifications

import (
	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
)

// Domain sentinel errors for in-app notification state. Codes
// pinned to internal/errors/codes.
var (
	// ErrTokenNotFound — 404. Returned when no FCM token row
	// exists for the (user, device) pair.
	ErrTokenNotFound = apperrors.NotFound(
		codes.PushTokenNotFound,
		"notification.token_not_found",
	)
	// ErrTokenInactive — 400. The token row exists but has
	// been disabled.
	ErrTokenInactive = apperrors.BadRequest(
		codes.PushTokenInactive,
		"notification.token_inactive",
	)
	// ErrDeviceMismatch — 403. The caller's session device
	// does not match the device the token is registered to.
	ErrDeviceMismatch = apperrors.Forbidden(
		codes.DeviceMismatch,
		"notification.device_mismatch",
	)
	// ErrPrefNotFound — 404. No notification preferences row
	// exists for the caller (defaults are normally seeded).
	ErrPrefNotFound = apperrors.NotFound(
		codes.NotificationPrefsNotFound,
		"notification.prefs_not_found",
	)
	// ErrInvalidTimezone — 400. The supplied IANA timezone is
	// not recognized.
	ErrInvalidTimezone = apperrors.BadRequest(
		codes.InvalidTimezone,
		"notification.invalid_timezone",
	)
	// ErrSendFailed — 500. Returned by the test-push endpoint
	// when the FCM client rejects the send.
	ErrSendFailed = apperrors.Internal(
		codes.PushSendFailed,
		"notification.send_failed",
	)
)
