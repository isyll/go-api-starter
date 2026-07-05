package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

const (
	TaskSendNotification     = "notification:send"
	TaskScheduleNotification = "notification:scheduled"
)

type Dispatcher interface {
	Send(ctx context.Context, event Event) error
	Schedule(ctx context.Context, event Event, at time.Time) error
	Close() error
}

type dispatcher struct {
	client *asynq.Client
	cfg    *config.NotificationsConfig
	logger *logger.Logger
}

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
		asynq.Queue(queueName(event.Priority)),
		asynq.MaxRetry(d.cfg.Worker.RetryMax),
		asynq.Timeout(30 * time.Second),
	}

	if event.IdempotencyKey != "" {
		opts = append(opts, asynq.TaskID(event.IdempotencyKey))
	}

	task := asynq.NewTask(TaskSendNotification, payload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		if isDuplicateEnqueue(err) {
			d.logger.Debug("Notification already enqueued; deduplicated",
				"event_type", event.Type,
				"idempotency_key", event.IdempotencyKey,
			)
			return nil
		}
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
		asynq.Queue(queueName(event.Priority)),
		asynq.MaxRetry(d.cfg.Worker.RetryMax),
		asynq.ProcessAt(at),
		asynq.Timeout(30 * time.Second),
	}

	if event.IdempotencyKey != "" {
		opts = append(opts, asynq.TaskID(event.IdempotencyKey))
	}

	task := asynq.NewTask(TaskScheduleNotification, payload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		if isDuplicateEnqueue(err) {
			d.logger.Debug("Notification already scheduled; deduplicated",
				"event_type", event.Type,
				"idempotency_key", event.IdempotencyKey,
			)
			return nil
		}
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

func isDuplicateEnqueue(err error) bool {
	return errors.Is(err, asynq.ErrDuplicateTask) ||
		errors.Is(err, asynq.ErrTaskIDConflict)
}
