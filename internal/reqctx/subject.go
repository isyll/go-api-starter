package reqctx

import "context"

type Role string

const (
	RoleAnonymous Role = "anonymous"
	RoleUser      Role = "user"
	RoleAdmin     Role = "admin"
)

// Subject is the authenticated principal for a request.
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

type subjectKey struct{}

func WithSubject(ctx context.Context, s Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

// SubjectFrom returns the request's authenticated subject (zero value if none).
func SubjectFrom(ctx context.Context) Subject {
	s, _ := ctx.Value(subjectKey{}).(Subject)
	return s
}
