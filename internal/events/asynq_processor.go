package events

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-api-starter/internal/reqctx"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// Processor decodes envelopes off events:dispatch and
// re-enters the bus to invoke registered async handlers.
type Processor struct {
	bus    *Bus
	logger *logger.Logger
}

// NewProcessor constructs a Processor backed by the supplied Bus.
// The worker's Bus is configured WITHOUT an AsyncDispatcher so
// handlers cannot accidentally re-publish from inside the worker.
func NewProcessor(bus *Bus, logx *logger.Logger) *Processor {
	return &Processor{bus: bus, logger: logx}
}

// ProcessTask decodes the envelope, allocates the
// concrete type via the registry, and dispatches.
// Unknown event types are skipped without error so an
// older worker does not crash on newly deployed events.
func (p *Processor) ProcessTask(
	ctx context.Context,
	t *asynq.Task,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			p.logger.Error(
				"event processor panic recovered",
				"task_type", t.Type(),
				"panic", r,
				"stack_trace", string(stack),
			)
			err = fmt.Errorf(
				"events: processor panic for task %s: %v",
				t.Type(), r,
			)
		}
	}()

	var env envelope
	if err := json.Unmarshal(t.Payload(), &env); err != nil {
		return fmt.Errorf("events: decode envelope: %w", err)
	}

	if env.RequestID != "" {
		ctx = reqctx.WithRequestID(ctx, env.RequestID)
	}

	p.logger.Debug(
		"processing event",
		"event", env.Type,
		"request_id", env.RequestID,
	)

	factory := FactoryFor(env.Type)
	if factory == nil {
		p.logger.Warn(
			"event type not registered, skipping",
			"event", env.Type,
		)
		return nil
	}

	evt := factory()
	if err := json.Unmarshal(env.Payload, evt); err != nil {
		return fmt.Errorf(
			"events: decode payload for %s: %w",
			env.Type, err,
		)
	}

	return p.bus.InvokeAsync(ctx, evt)
}
