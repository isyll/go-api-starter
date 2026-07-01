package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-api-starter/pkg/config"
	"github.com/isyll/go-api-starter/pkg/logger"
)

const (
	// TaskSendNotification is the Asynq task type for
	// immediate push notification delivery.
	TaskSendNotification = "notification:send"
	// TaskScheduleNotification is the Asynq task type for
	// time-delayed notification delivery.
	TaskScheduleNotification = "notification:scheduled"
)

// Dispatcher handles push-notification task queuing. Callers enqueue
// an Event and the worker delivers the FCM payload to the target
// device. All queueing is fire-and-forget; handlers should never
// await delivery confirmation in the request path.
type Dispatcher interface {
	// Send enqueues an immediate push-notification task. The event's
	// IdempotencyKey, if set, prevents duplicate delivery within the
	// asynq unique window.
	Send(ctx context.Context, event Event) error
	// Schedule enqueues a push-notification task for delivery at at.
	Schedule(ctx context.Context, event Event, at time.Time) error
	// Close releases the underlying Asynq client.
	Close() error
}

type dispatcher struct {
	client *asynq.Client
	cfg    *config.NotificationsConfig
	logger *logger.Logger
}

// NewDispatcher constructs a notifications Dispatcher that connects
// to the Asynq Redis backend at redisAddr.
func NewDispatcher(
	redisAddr string,
	redisPassword string,
	cfg *config.NotificationsConfig,
	logx *logger.Logger,
) Dispatcher {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	return &dispatcher{
		client: client,
		cfg:    cfg,
		logger: logx,
	}
}

func (d *dispatcher) Send(ctx context.Context, event Event) error {
	if event.Priority == "" {
		event.Priority = PriorityNormal
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue(string(event.Priority)),
		asynq.MaxRetry(d.cfg.Worker.RetryMax),
		asynq.Timeout(30 * time.Second),
	}

	if event.IdempotencyKey != "" {
		opts = append(opts, asynq.TaskID(event.IdempotencyKey))
	}

	task := asynq.NewTask(TaskSendNotification, payload, opts...)

	info, err := d.client.Enqueue(task)
	if err != nil {
		d.logger.Error("Failed to enqueue notification",
			"event_type", event.Type,
			"user_id", event.UserID,
			"error", err,
		)
		return fmt.Errorf("failed to enqueue notification: %w", err)
	}

	d.logger.Debug("Notification enqueued",
		"task_id", info.ID,
		"queue", info.Queue,
		"event_type", event.Type,
		"user_id", event.UserID,
	)

	return nil
}

func (d *dispatcher) Schedule(
	ctx context.Context,
	event Event,
	at time.Time,
) error {
	if event.Priority == "" {
		event.Priority = PriorityNormal
	}
	event.ScheduledAt = &at

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue(string(event.Priority)),
		asynq.MaxRetry(d.cfg.Worker.RetryMax),
		asynq.ProcessAt(at),
		asynq.Timeout(30 * time.Second),
	}

	if event.IdempotencyKey != "" {
		opts = append(opts, asynq.TaskID(event.IdempotencyKey))
	}

	task := asynq.NewTask(TaskScheduleNotification, payload, opts...)

	info, err := d.client.Enqueue(task)
	if err != nil {
		d.logger.Error("Failed to schedule notification",
			"event_type", event.Type,
			"user_id", event.UserID,
			"scheduled_at", at,
			"error", err,
		)
		return fmt.Errorf("failed to schedule notification: %w", err)
	}

	d.logger.Debug("Notification scheduled",
		"task_id", info.ID,
		"queue", info.Queue,
		"event_type", event.Type,
		"user_id", event.UserID,
		"scheduled_at", at,
	)

	return nil
}

func (d *dispatcher) Close() error {
	return d.client.Close()
}
