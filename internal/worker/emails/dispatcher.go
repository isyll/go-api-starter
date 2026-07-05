package emails

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
	TaskSendEmail      = "email:send"
	TaskScheduledEmail = "email:scheduled"
	TaskBulkEmail      = "email:bulk"
)

type Dispatcher interface {
	Send(ctx context.Context, email *Email) error
	SendBulk(ctx context.Context, emails []*Email) error
	Schedule(ctx context.Context, email *Email, at time.Time) error
	Close() error
}

type dispatcher struct {
	client *asynq.Client
	cfg    *config.EmailConfig
	logger *logger.Logger
}

func NewDispatcher(
	redisAddr string,
	redisPassword string,
	cfg *config.EmailConfig,
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

func (d *dispatcher) Send(ctx context.Context, email *Email) error {
	if email.Priority == "" {
		email.Priority = PriorityNormal
	}

	if email.Language == "" {
		email.Language = d.cfg.Email.Templates.DefaultLanguage
	}

	payload, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue(queueName(email.Priority)),
		asynq.MaxRetry(d.cfg.Email.Worker.RetryMax),
		asynq.Timeout(60 * time.Second),
	}

	if email.IdempotencyKey != "" {
		opts = append(opts, asynq.TaskID(email.IdempotencyKey))
	}

	task := asynq.NewTask(TaskSendEmail, payload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		if isDuplicateEnqueue(err) {
			d.logger.Debug("Email already enqueued; deduplicated",
				"type", email.Type,
				"idempotency_key", email.IdempotencyKey,
			)
			return nil
		}
		d.logger.Error("Failed to enqueue email",
			"type", email.Type,
			"to", email.To,
			"error", err,
		)
		return fmt.Errorf("failed to enqueue email: %w", err)
	}

	d.logger.Debug("Email enqueued",
		"task_id", info.ID,
		"queue", info.Queue,
		"type", email.Type,
		"to", email.To,
	)

	return nil
}

func (d *dispatcher) SendBulk(
	ctx context.Context,
	emails []*Email,
) error {
	if len(emails) == 0 {
		return nil
	}

	batchSize := d.cfg.Email.Batch.MaxSize
	for i := 0; i < len(emails); i += batchSize {
		end := i + batchSize
		if end > len(emails) {
			end = len(emails)
		}

		batch := emails[i:end]
		payload, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("failed to marshal email batch: %w", err)
		}

		opts := []asynq.Option{
			asynq.Queue(queueName(PriorityLow)),
			asynq.MaxRetry(d.cfg.Email.Worker.RetryMax),
			asynq.Timeout(5 * time.Minute),
		}

		task := asynq.NewTask(TaskBulkEmail, payload, opts...)

		info, err := d.client.EnqueueContext(ctx, task)
		if err != nil {
			d.logger.Error("Failed to enqueue bulk email",
				"batch_size", len(batch),
				"error", err,
			)
			return fmt.Errorf("failed to enqueue bulk email: %w", err)
		}

		d.logger.Debug("Bulk email enqueued",
			"task_id", info.ID,
			"batch_size", len(batch),
		)
	}

	return nil
}

func (d *dispatcher) Schedule(
	ctx context.Context,
	email *Email,
	at time.Time,
) error {
	if email.Priority == "" {
		email.Priority = PriorityNormal
	}
	email.ScheduledAt = &at

	if email.Language == "" {
		email.Language = d.cfg.Email.Templates.DefaultLanguage
	}

	payload, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue(queueName(email.Priority)),
		asynq.MaxRetry(d.cfg.Email.Worker.RetryMax),
		asynq.ProcessAt(at),
		asynq.Timeout(60 * time.Second),
	}

	if email.IdempotencyKey != "" {
		opts = append(opts, asynq.TaskID(email.IdempotencyKey))
	}

	task := asynq.NewTask(TaskScheduledEmail, payload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		if isDuplicateEnqueue(err) {
			d.logger.Debug("Email already scheduled; deduplicated",
				"type", email.Type,
				"idempotency_key", email.IdempotencyKey,
			)
			return nil
		}
		d.logger.Error("Failed to schedule email",
			"type", email.Type,
			"to", email.To,
			"scheduled_at", at,
			"error", err,
		)
		return fmt.Errorf("failed to schedule email: %w", err)
	}

	d.logger.Debug("Email scheduled",
		"task_id", info.ID,
		"type", email.Type,
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
