// Package authz holds the authenticated principal (Subject) and its
// role, carried through the request context.
package authz

// Role is the user's functional role.
type Role string

const (
	// RoleAnonymous is an unauthenticated caller.
	RoleAnonymous Role = "anonymous"
	// RoleUser is a normal end user.
	RoleUser Role = "user"
	// RoleAdmin is a back-office operator.
	RoleAdmin Role = "admin"
)

// Subject is the authenticated principal for the current request. It is
// populated by the auth interceptor from the validated access token and
// session.
type Subject struct {
	UserID    int64
	Role      Role
	SessionID int64
	DeviceID  string
	IsAdmin   bool
}

// IsAnonymous reports whether the subject is unauthenticated.
func (s Subject) IsAnonymous() bool {
	return s.UserID == 0 && !s.IsAdmin
}

// HasRole reports whether the subject may act in role r. Admins satisfy
// every role check.
func (s Subject) HasRole(r Role) bool {
	if s.IsAdmin {
		return true
	}
	return s.Role == r
}
