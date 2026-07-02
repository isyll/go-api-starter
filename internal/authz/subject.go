// Package authz holds the authenticated principal and its role.
package authz

type Role string

const (
	RoleAnonymous Role = "anonymous"
	RoleUser      Role = "user"
	RoleAdmin     Role = "admin"
)

type Subject struct {
	UserID    int64
	Role      Role
	SessionID int64
	DeviceID  string
	IsAdmin   bool
}

func (s Subject) IsAnonymous() bool {
	return s.UserID == 0 && !s.IsAdmin
}

func (s Subject) HasRole(r Role) bool {
	if s.IsAdmin {
		return true
	}
	return s.Role == r
}
