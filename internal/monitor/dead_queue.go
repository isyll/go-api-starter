package monitor

import (
	"context"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-api-starter/internal/metrics"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// DeadQueueMonitor polls Asynq queue stats at the given
// interval and updates the app_asynq_dead_queue_size gauge
// for each queue. Call Run in a goroutine; it blocks until ctx
// is canceled.
type DeadQueueMonitor struct {
	inspector *asynq.Inspector
	interval  time.Duration
	queues    []string
	logger    *logger.Logger
}

func NewDeadQueueMonitor(
	redisAddr string,
	redisPassword string,
	interval time.Duration,
	queues []string,
	logx *logger.Logger,
) *DeadQueueMonitor {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	return &DeadQueueMonitor{
		inspector: inspector,
		interval:  interval,
		queues:    queues,
		logger:    logx,
	}
}

func (m *DeadQueueMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collect()
		}
	}
}

func (m *DeadQueueMonitor) collect() {
	for _, q := range m.queues {
		info, err := m.inspector.GetQueueInfo(q)
		if err != nil {
			m.logger.Warn(
				"dead queue monitor: failed to get queue info",
				"queue", q,
				"error", err,
			)
			continue
		}

		metrics.AsynqDeadQueueSize.
			WithLabelValues(q).
			Set(float64(info.Archived))

		if info.Archived > 0 {
			m.logger.Warn(
				"archived (dead) queue has tasks — manual review required",
				"queue",
				q,
				"archived_count",
				info.Archived,
			)
		}
	}
}
