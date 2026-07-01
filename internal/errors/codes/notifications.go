package codes

// Notifications-domain codes. /my/devices/tokens, notification
// preferences, and the test-push diagnostic endpoint.

var (
	// PushTokenNotFound — 404. No FCM token row for the (user, device) pair.
	PushTokenNotFound = register("PUSH_TOKEN_NOT_FOUND")

	// PushTokenInactive — 400. Token row exists but is marked inactive.
	PushTokenInactive = register("PUSH_TOKEN_INACTIVE")

	// DeviceMismatch — 403. Caller's session device != token's registered device.
	DeviceMismatch = register("DEVICE_MISMATCH")

	// NotificationPrefsNotFound — 404. Preferences row missing (normally seeded on signup).
	NotificationPrefsNotFound = register("NOTIFICATION_PREFS_NOT_FOUND")

	// PushSendFailed — 500. Test-push endpoint and FCM client rejected the send.
	PushSendFailed = register("PUSH_SEND_FAILED")
)
