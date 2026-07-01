// Package grpc holds the gRPC server, interceptors, and the service
// implementations that adapt domain services to the generated API.
package grpc

import (
	"context"

	"github.com/isyll/go-api-starter/internal/models"
)

type ctxKey int

const (
	userKey ctxKey = iota
	sessionIDKey
)

func withUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// userFrom returns the authenticated user set by the auth interceptor.
func userFrom(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(userKey).(*models.User)
	return u, ok
}

func withSessionID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

func sessionIDFrom(ctx context.Context) int64 {
	id, _ := ctx.Value(sessionIDKey).(int64)
	return id
}
