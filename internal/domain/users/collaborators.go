package users

import "context"

// SessionRevoker revokes all sessions for a user. Implemented by the
// auth service; account deletion uses it to sign the user out
// everywhere. Declared here to avoid importing the auth package.
type SessionRevoker interface {
	RevokeAllSessions(ctx context.Context, userID int64, reason string) error
}
