package persistence

import (
	"context"
	"sync/atomic"
)

type queryCounter struct {
	n atomic.Int64
}

type queryCounterKey struct{}

// WithQueryCounter injects a fresh query counter into ctx.
// The GORM callback registered by RegisterQueryCallbacks
// increments the counter on every DB statement that executes
// with the returned context.
func WithQueryCounter(ctx context.Context) context.Context {
	return context.WithValue(
		ctx, queryCounterKey{}, &queryCounter{},
	)
}

// IncrQueryCounter increments the per-request DB statement count.
// No-ops when no counter has been injected (e.g. background tasks).
func IncrQueryCounter(ctx context.Context) {
	if c, ok := ctx.Value(queryCounterKey{}).(*queryCounter); ok {
		c.n.Add(1)
	}
}

// QueryCount returns the number of DB statements executed on ctx.
func QueryCount(ctx context.Context) int64 {
	if c, ok := ctx.Value(queryCounterKey{}).(*queryCounter); ok {
		return c.n.Load()
	}
	return 0
}
