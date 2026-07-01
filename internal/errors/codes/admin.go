package codes

// Admin-API codes. Auth codes keep ADMIN_ in the wire form so
// the back-office telemetry can isolate admin-side failures
// without URL parsing. Generic misses reuse cross-domain codes.

var (
	// AdminMissingToken — 401. Service JWT absent or Authorization header malformed.
	AdminMissingToken = register("ADMIN_MISSING_TOKEN")

	// AdminTokenExpired — 401. Service JWT expired; admin frontend silently refreshes.
	AdminTokenExpired = register("ADMIN_TOKEN_EXPIRED")

	// InvalidAdminToken — 401. Service JWT failed signature validation.
	InvalidAdminToken = register("INVALID_ADMIN_TOKEN")

	// AdminForbiddenSource — 403. JWT service_source claim != "admin_backoffice".
	AdminForbiddenSource = register("ADMIN_FORBIDDEN_SOURCE")

	// AdminSessionInvalidated — 401. Admin's Redis invalidation marker is set.
	AdminSessionInvalidated = register("ADMIN_SESSION_INVALIDATED")

	// InsufficientPermissions — 403. Admin's permissions missing the required entry.
	InsufficientPermissions = register("INSUFFICIENT_PERMISSIONS")

	// CountryScopeViolation — 403. Country not in admin's country_codes (global ['*'] passes).
	CountryScopeViolation = register("COUNTRY_SCOPE_VIOLATION")

	// AdminTrackingRoomNotFound — 404. Trip tracking room is not currently open.
	AdminTrackingRoomNotFound = register(
		"ADMIN_TRACKING_ROOM_NOT_FOUND",
	)
)
