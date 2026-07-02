package events

import (
	"context"
	"fmt"

	"github.com/isyll/go-api-starter/internal/metrics"
)

func (r *OutboxRepository) UpdateMetrics(
	ctx context.Context,
) error {
	type row struct {
		Pending   int64
		Exhausted int64
		LagSecs   float64
	}

	var result row
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (
				WHERE processed_at IS NULL
				  AND retry_count < ?
			) AS pending,
			COUNT(*) FILTER (
				WHERE retry_count >= ?
			) AS exhausted,
			COALESCE(
				EXTRACT(EPOCH FROM (
					now() - MIN(created_at)
				)) FILTER (
					WHERE processed_at IS NULL
					  AND retry_count < ?
				),
				0
			) AS lag_secs
		FROM events.outbox
	`, outboxMaxRetry, outboxMaxRetry, outboxMaxRetry).
		Scan(&result).Error
	if err != nil {
		return fmt.Errorf("outbox metrics: %w", err)
	}

	metrics.OutboxPendingRows.Set(float64(result.Pending))
	metrics.OutboxExhaustedRows.Set(float64(result.Exhausted))
	metrics.OutboxDrainLagSeconds.Set(result.LagSecs)

	var dlDepth int64
	if err := r.db.WithContext(ctx).
		Raw(`SELECT COUNT(*) FROM events.outbox_dead_letters`).
		Scan(&dlDepth).Error; err != nil {
		return fmt.Errorf("outbox dead-letter depth: %w", err)
	}
	metrics.OutboxDeadLetterDepth.Set(float64(dlDepth))

	return nil
}
