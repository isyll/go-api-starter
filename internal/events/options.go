package events

import (
	"time"

	"github.com/hibiken/asynq"
)

// SubscribeOption customizes a Subscribe / SubscribeAsync
// registration: criticality, queue name, idempotency window,
// task-ID derivation.
type SubscribeOption func(*subConfig)

type subConfig struct {
	critical  bool
	asyncOpts []asynq.Option
	taskIDFn  func(Event) string
}

func newSubConfig(opts []SubscribeOption) subConfig {
	cfg := subConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// WithCritical surfaces a sync handler's error back to
// the publisher. Reserve for handlers whose failure must
// fail the request.
func WithCritical() SubscribeOption {
	return func(c *subConfig) { c.critical = true }
}

// WithQueue routes async deliveries of this subscription to the
// named Asynq queue. Useful when distinct event handlers must be
// drained by separate workers (e.g. notifications vs emails).
func WithQueue(name string) SubscribeOption {
	return func(c *subConfig) {
		c.asyncOpts = append(c.asyncOpts, asynq.Queue(name))
	}
}

// WithUniqueWindow declares an idempotency window on
// the Asynq task. Combine with WithTaskIDFn.
func WithUniqueWindow(d time.Duration) SubscribeOption {
	return func(c *subConfig) {
		c.asyncOpts = append(c.asyncOpts, asynq.Unique(d))
	}
}

// WithTaskIDFn derives the Asynq task ID from the
// published event at enqueue time. This prevents
// duplicate tasks when Asynq retries the
// events:dispatch envelope after a transient failure.
// Combine with WithUniqueWindow to set the dedup window.
func WithTaskIDFn(fn func(Event) string) SubscribeOption {
	return func(c *subConfig) { c.taskIDFn = fn }
}
