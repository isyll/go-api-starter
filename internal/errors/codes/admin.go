package codes

// Admin-API error codes.
var (
	// AdminRequired - 403. The caller is not an admin.
	AdminRequired = register("ADMIN_REQUIRED")

	// InsufficientPermissions - 403. The admin lacks the required permission.
	InsufficientPermissions = register("INSUFFICIENT_PERMISSIONS")
)
