package authz

// Role is the user's functional role within the platform.
// Roles are coarse-grained capabilities; per-resource authorization
// is decided by Policy.Can on top of the role.
type Role string

const (
	// RoleAnonymous indicates the user is not authenticated.
	RoleAnonymous Role = "anonymous"
	// RoleDriver indicates a user who may publish trips and accept
	// bookings.
	RoleDriver Role = "driver"
	// RolePassenger indicates a user who may book trips published by
	// drivers.
	RolePassenger Role = "passenger"
	// RoleBoth indicates the user may act as driver or passenger.
	RoleBoth Role = "both"
	// RoleAdmin indicates a back-office operator with privileged
	// access via the admin surface (separate auth, see
	// internal/middleware/admin).
	RoleAdmin Role = "admin"
)

// Subject is the authenticated principal for the current request.
// It is the authoritative source for all authorization decisions and
// is populated by AuthRequired / AuthOptional middleware from the
// validated OAT (opaque access token) session.
type Subject struct {
	// UserID is the decoded DB primary key. Never expose in responses;
	// always pass through idenc.IDEncoder at the response boundary.
	UserID int64
	// Role is the user's functional role (driver, passenger, both or admin).
	Role Role
	// SessionID identifies the OAT session row this request came from.
	SessionID int64
	// DeviceID is the trusted device identifier derived from the
	// session record, never trusted from request headers.
	DeviceID string
	// Country is the ISO 3166-1 alpha-2 code from the user's profile.
	// Used for country-scoped operational availability checks.
	Country string
	// IsAdmin reports whether the request came in through the admin
	// surface; admin policy branches must always allow first.
	IsAdmin bool
}

// IsAnonymous reports whether the subject is unauthenticated.
func (s Subject) IsAnonymous() bool {
	return s.UserID == 0 && !s.IsAdmin
}

// HasRole reports whether the subject may act in role r.
// Admins satisfy every role check.
func (s Subject) HasRole(r Role) bool {
	if s.IsAdmin {
		return true
	}
	return s.Role == r ||
		(s.Role == RoleBoth &&
			(r == RoleDriver || r == RolePassenger))
}
