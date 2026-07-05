package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "app"

	subsystemAdmin  = "admin"
	subsystemEvents = "events"
	subsystemWorker = "worker"
	subsystemCache  = "cache"
	subsystemOutbox = "outbox"

	labelEventType = "event_type"
)

var (
	EventsPublishedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemEvents,
			Name:      "published_total",
			Help: "Total number of events published to " +
				"the bus.",
		},
		[]string{labelEventType},
	)

	EventsEnqueueFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemEvents,
			Name:      "enqueue_failures_total",
			Help: "Total number of async event enqueue " +
				"failures.",
		},
		[]string{labelEventType},
	)

	EventsEnqueueDedupedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemEvents,
			Name:      "enqueue_deduped_total",
			Help: "Redeliveries skipped because the task " +
				"was already enqueued (unique or task id match).",
		},
		[]string{labelEventType},
	)

	EventsHandlerDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemEvents,
			Name:      "handler_duration_seconds",
			Help: "Duration of event handler execution " +
				"in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type", "handler", "kind"},
	)

	WorkerTasksProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemWorker,
			Name:      "tasks_processed_total",
			Help: "Total number of Asynq tasks processed " +
				"by domain workers.",
		},
		[]string{subsystemWorker, "task_type", "status"},
	)

	WorkerTaskDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemWorker,
			Name:      "task_duration_seconds",
			Help: "Duration of Asynq task processing in " +
				"seconds.",
			Buckets: []float64{
				0.01, 0.05, 0.1, 0.25, 0.5,
				1, 2.5, 5, 10, 30,
			},
		},
		[]string{subsystemWorker, "task_type"},
	)

	WorkerPanicsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemWorker,
			Name:      "panics_total",
			Help: "Total number of panics recovered in " +
				"worker processors.",
		},
		[]string{subsystemWorker},
	)

	CacheInvalidationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemCache,
			Name:      "invalidations_total",
			Help: "Total number of cache tags invalidated " +
				"by event handlers.",
		},
		[]string{labelEventType},
	)

	CacheInvalidationFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemCache,
			Name:      "invalidation_failed_total",
			Help: "Total number of cache invalidation calls " +
				"that returned an error.",
		},
		[]string{"event_type", "tags"},
	)

	CacheSingleflightCollapsedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemCache,
			Name:      "singleflight_collapsed_total",
			Help: "Number of cache fetches that joined an " +
				"in-flight singleflight call instead of hitting " +
				"the backing store directly.",
		},
		[]string{"key_kind"},
	)

	AsynqQueueDepth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "asynq",
			Name:      "queue_depth",
			Help: "Tasks per queue and state (pending, active, " +
				"scheduled, retry, archived).",
		},
		[]string{"queue", "state"},
	)

	AsynqQueueLatencySeconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "asynq",
			Name:      "queue_latency_seconds",
			Help: "Age of the oldest pending task per queue; " +
				"sustained growth means the queue is falling behind.",
		},
		[]string{"queue"},
	)

	OutboxPendingRows = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemOutbox,
			Name:      "pending_rows",
			Help: "Outbox rows awaiting processing " +
				"(retry_count < max).",
		},
	)

	OutboxExhaustedRows = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemOutbox,
			Name:      "exhausted_rows",
			Help: "Outbox rows that exceeded the " +
				"maximum retry count.",
		},
	)

	OutboxDrainLagSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemOutbox,
			Name:      "drain_lag_seconds",
			Help: "Age of the oldest pending outbox " +
				"row in seconds.",
		},
	)

	OutboxMarkFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemOutbox,
			Name:      "mark_failures_total",
			Help: "Number of MarkProcessed/MarkFailed " +
				"calls that returned an error.",
		},
		[]string{"op"},
	)

	OutboxDeadLetterTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemOutbox,
			Name:      "dead_letter_total",
			Help: "Number of outbox rows moved to the " +
				"dead-letter table.",
		},
		[]string{"reason"},
	)

	OutboxDeadLetterDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemOutbox,
			Name:      "dead_letter_depth",
			Help: "Current number of rows in the outbox " +
				"dead-letter table awaiting operator action.",
		},
	)

	MaintenanceRowsDeletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "maintenance",
			Name:      "rows_deleted_total",
			Help: "Rows removed by the worker's periodic " +
				"retention sweeps, by job.",
		},
		[]string{"job"},
	)

	AuditLogWriteFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: subsystemAdmin,
			Name:      "audit_log_write_failures_total",
			Help: "Total failures writing admin audit log " +
				"entries, by failure reason.",
		},
		[]string{"reason"},
	)
)
