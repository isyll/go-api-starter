package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-grpc-starter/internal/reqctx"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

const TaskTypeDispatch = "events:dispatch"

type envelope struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	RequestID string          `json:"request_id,omitempty"`
}

type AsynqDispatcher struct {
	client *asynq.Client
	logger *logger.Logger
}

func NewAsynqDispatcher(
	client *asynq.Client,
	logx *logger.Logger,
) *AsynqDispatcher {
	return &AsynqDispatcher{client: client, logger: logx}
}

func (d *AsynqDispatcher) Enqueue(
	ctx context.Context,
	evt Event,
	opts []asynq.Option,
) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("events: marshal payload: %w", err)
	}
	body, err := json.Marshal(envelope{
		Type:      evt.EventType(),
		Payload:   payload,
		RequestID: reqctx.RequestIDFromContext(ctx),
	})
	if err != nil {
		return fmt.Errorf("events: marshal envelope: %w", err)
	}

	task := asynq.NewTask(TaskTypeDispatch, body, opts...)
	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf(
			"events: enqueue %s: %w",
			evt.EventType(), err,
		)
	}

	d.logger.Debug("event task enqueued",
		"task_id", info.ID,
		"queue", info.Queue,
		"event", evt.EventType(),
		"request_id", reqctx.RequestIDFromContext(ctx),
	)
	return nil
}

func (d *AsynqDispatcher) Close() error {
	return d.client.Close()
}
