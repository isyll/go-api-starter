// Package handlers holds the reactions wired to bus event.
package handlers

import (
	"context"

	"github.com/isyll/go-grpc-starter/internal/event"
	"github.com/isyll/go-grpc-starter/internal/metrics"
	"github.com/isyll/go-grpc-starter/internal/platform/cache"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

type CacheInvalidator struct {
	cache  *cache.CacheManager
	logger *logger.Logger
}

func NewCacheInvalidator(
	c *cache.CacheManager,
	logx *logger.Logger,
) *CacheInvalidator {
	return &CacheInvalidator{cache: c, logger: logx}
}

func (h *CacheInvalidator) OnUserAccountDeleted(
	ctx context.Context,
	evt *event.UserAccountDeleted,
) error {
	return h.invalidate(ctx, evt.EventType(), cache.UserTag(evt.UserID))
}

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
