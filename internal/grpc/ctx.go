// Package grpc holds the gRPC server, interceptors, and services.
package grpc

import (
	"context"

	"github.com/isyll/go-grpc-starter/internal/users"
)

type ctxKey int

const (
	userKey ctxKey = iota
	sessionIDKey
)

func withUser(ctx context.Context, u *users.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func userFrom(ctx context.Context) (*users.User, bool) {
	u, ok := ctx.Value(userKey).(*users.User)
	return u, ok
}

func withSessionID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

func sessionIDFrom(ctx context.Context) int64 {
	id, _ := ctx.Value(sessionIDKey).(int64)
	return id
}
