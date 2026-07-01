// Package reqctx carries the request ID across context boundaries so
// background workers and event handlers can log the same correlation
// ID as the originating request.
package reqctx

import "context"

type ctxKey string

const requestIDKey ctxKey = "request_id"

// WithRequestID returns a copy of ctx carrying the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext returns the request ID on ctx, or "".
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
