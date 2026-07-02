package persistence

import (
	"context"
	"sync/atomic"
)

type queryCounter struct {
	n atomic.Int64
}

type queryCounterKey struct{}

func WithQueryCounter(ctx context.Context) context.Context {
	return context.WithValue(
		ctx, queryCounterKey{}, &queryCounter{},
	)
}

func IncrQueryCounter(ctx context.Context) {
	if c, ok := ctx.Value(queryCounterKey{}).(*queryCounter); ok {
		c.n.Add(1)
	}
}

func QueryCount(ctx context.Context) int64 {
	if c, ok := ctx.Value(queryCounterKey{}).(*queryCounter); ok {
		return c.n.Load()
	}
	return 0
}
