package users

import "context"

type SessionRevoker interface {
	RevokeAllSessions(ctx context.Context, userID int64, reason string) error
}
