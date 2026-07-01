package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-api-starter/internal/reqctx"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// TaskTypeDispatch is the single Asynq task name for
// every async event - the worker decodes the envelope to
// learn which handlers to run.
const TaskTypeDispatch = "events:dispatch"

type envelope struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	RequestID string          `json:"request_id,omitempty"`
}

// AsynqDispatcher is the production AsyncDispatcher: it marshals
// the published Event into a JSON envelope and enqueues a single
// TaskTypeDispatch Asynq task for the event-dispatcher worker to
// consume. The same Bus configuration is rebuilt on the worker so
// async handlers run with the identical registrations.
type AsynqDispatcher struct {
	client *asynq.Client
	logger *logger.Logger
}

// NewAsynqDispatcher constructs an AsynqDispatcher around an
// existing Asynq client. The client's lifecycle is owned by the
// dispatcher; call Close to release it.
func NewAsynqDispatcher(
	client *asynq.Client,
	logx *logger.Logger,
) *AsynqDispatcher {
	return &AsynqDispatcher{client: client, logger: logx}
}

// Enqueue marshals evt into the dispatcher envelope (event type +
// payload + request ID) and writes a TaskTypeDispatch task to Asynq
// with the supplied per-subscription options (queue, idempotency
// window, task ID).
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

// Close releases the underlying Asynq client. Idempotent.
func (d *AsynqDispatcher) Close() error {
	return d.client.Close()
}
