package monitor

import (
	"context"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-grpc-starter/internal/metrics"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

// QueueMonitor polls Asynq queue depth per state into Prometheus gauges.
type QueueMonitor struct {
	inspector *asynq.Inspector
	interval  time.Duration
	queues    []string
	logger    *logger.Logger
}

func NewQueueMonitor(
	redisAddr string,
	redisPassword string,
	interval time.Duration,
	queues []string,
	logx *logger.Logger,
) *QueueMonitor {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	return &QueueMonitor{
		inspector: inspector,
		interval:  interval,
		queues:    queues,
		logger:    logx,
	}
}

func (m *QueueMonitor) Run(ctx context.Context) {
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

func (m *QueueMonitor) collect() {
	for _, q := range m.queues {
		info, err := m.inspector.GetQueueInfo(q)
		if err != nil {
			// Queues exist only after their first task; stay quiet until then.
			m.logger.Debug(
				"queue monitor: queue info unavailable",
				"queue", q,
				"error", err,
			)
			continue
		}

		depth := metrics.AsynqQueueDepth
		depth.WithLabelValues(q, "pending").Set(float64(info.Pending))
		depth.WithLabelValues(q, "active").Set(float64(info.Active))
		depth.WithLabelValues(q, "scheduled").Set(float64(info.Scheduled))
		depth.WithLabelValues(q, "retry").Set(float64(info.Retry))
		depth.WithLabelValues(q, "archived").Set(float64(info.Archived))

		metrics.AsynqQueueLatencySeconds.
			WithLabelValues(q).
			Set(info.Latency.Seconds())

		if info.Archived > 0 {
			m.logger.Warn(
				"archived (dead) queue has tasks - manual review required",
				"queue", q,
				"archived_count", info.Archived,
			)
		}
	}
}
