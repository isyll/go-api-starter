// Package handlers wires concrete reactions to bus events: cache
// invalidation, audit-log persistence, and auth-attempt recording.
// Each exported method is registered against the *events.Bus in the
// app wiring. Cache invalidation runs sync; anything doing I/O runs
// async with an idempotency key so Asynq redelivery is safe.
package handlers

import (
	"context"

	"github.com/isyll/go-api-starter/internal/events"
	"github.com/isyll/go-api-starter/internal/infra/cache"
	"github.com/isyll/go-api-starter/internal/metrics"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// CacheInvalidator maps a domain change to the cache tags it drops.
type CacheInvalidator struct {
	cache  *cache.CacheManager
	logger *logger.Logger
}

// NewCacheInvalidator builds a CacheInvalidator over the shared
// CacheManager.
func NewCacheInvalidator(
	c *cache.CacheManager,
	logx *logger.Logger,
) *CacheInvalidator {
	return &CacheInvalidator{cache: c, logger: logx}
}

// OnUserAccountDeleted drops everything cached for a deleted account.
func (h *CacheInvalidator) OnUserAccountDeleted(
	ctx context.Context,
	evt *events.UserAccountDeleted,
) error {
	return h.invalidate(ctx, evt.EventType(), cache.UserTag(evt.UserID))
}

// invalidate drops the given tags and records the outcome as a
// metric. Failures are logged, not fatal: stale entries expire by TTL.
func (h *CacheInvalidator) invalidate(
	ctx context.Context,
	eventType string,
	tags ...string,
) error {
	if err := h.cache.InvalidateByTags(ctx, tags...); err != nil {
		bucket := "single"
		if len(tags) > 1 {
			bucket = "multi"
		}
		metrics.CacheInvalidationFailedTotal.
			WithLabelValues(eventType, bucket).
			Inc()
		h.logger.Warn(
			"cache invalidation failed",
			"event", eventType, "tags", tags, "error", err,
		)
		return err
	}
	metrics.CacheInvalidationsTotal.WithLabelValues(eventType).Inc()
	return nil
}
