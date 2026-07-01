package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/hibiken/asynq"

	"github.com/isyll/go-api-starter/internal/metrics"
	"github.com/isyll/go-api-starter/internal/persistence"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// ErrEnqueueFailed is returned by Publish when at least
// one async subscription failed to enqueue its task.
// The outbox (if configured) will retry those tasks.
// This error does NOT indicate that sync handlers failed
// — check WithCritical for that.
var ErrEnqueueFailed = errors.New(
	"one or more async event enqueues failed",
)

// AsyncDispatcher is the seam for async fan-out;
// production uses AsynqDispatcher, tests can swap it.
type AsyncDispatcher interface {
	Enqueue(
		ctx context.Context,
		evt Event,
		opts []asynq.Option,
	) error
}

// Bus is the in-process dispatch table. The same Bus
// instance is rebuilt in the event-dispatcher worker so
// async handlers there call the exact same registered
// functions.
//
// Publish is tx-aware: when the caller passes a context
// carrying a *gorm.DB transaction (via
// persistence.Manager.WithTx), the bus writes the outbox
// row INSIDE that transaction and returns immediately.
// Sync handlers and async enqueue happen later when the
// drain goroutine re-publishes the row in drainCtx after
// the tx has committed. This is the single mandatory
// pattern for any event that follows a domain mutation —
// it guarantees the outbox row and the mutation share
// commit fate.
//
// When Publish is called WITHOUT a transaction in the
// context — the sanctioned shape for non-mutation events
// (e.g. system observability) — the bus writes the outbox
// row in its own one-statement tx, runs sync handlers, and
// attempts async enqueue. The drain still catches any
// enqueue failure.
type Bus struct {
	mu     sync.RWMutex
	subs   map[string][]subscription
	asynq  AsyncDispatcher
	outbox *OutboxRepository
	logger *logger.Logger
	// drainKick signals the drain goroutine to run one
	// extra iteration immediately. The retry watchdog uses
	// this to cut Redis-recovery latency from up-to-30s to
	// under 2s after the first successful Asynq enqueue
	// following an outage.
	drainKick chan struct{}

	// drainBreakerFailures counts consecutive drain passes
	// where every fetched row failed to publish. When the
	// count crosses drainBreakerThreshold, drainOnce sleeps
	// drainBreakerCooldown before retrying. A successful
	// pass resets the counter. Mutated by drainOnce only;
	// no lock needed because drainOnce runs sequentially on
	// a single goroutine.
	drainBreakerFailures int
	drainBreakerOpenedAt time.Time
}

// drainKey is a context key used to signal that Publish
// is being called from the outbox drain goroutine so it
// must not write a new outbox row.
type drainKey struct{}

type subscription struct {
	name      string
	invoke    func(context.Context, Event) error
	isAsync   bool
	asyncOpts []asynq.Option
	taskIDFn  func(Event) string
	critical  bool
}

// New creates a Bus. Pass nil for asynq to disable
// async fan-out (used in the worker process to prevent
// re-publish loops).
func New(dispatcher AsyncDispatcher, logx *logger.Logger) *Bus {
	return &Bus{
		subs:      map[string][]subscription{},
		asynq:     dispatcher,
		logger:    logx,
		drainKick: make(chan struct{}, 1),
	}
}

// NewWithOutbox creates a Bus with outbox-backed retry
// for failed Asynq enqueues. The drain goroutine must be
// started separately via Bus.DrainOutbox.
func NewWithOutbox(
	dispatcher AsyncDispatcher,
	outbox *OutboxRepository,
	logx *logger.Logger,
) *Bus {
	return &Bus{
		subs:      map[string][]subscription{},
		asynq:     dispatcher,
		outbox:    outbox,
		logger:    logx,
		drainKick: make(chan struct{}, 1),
	}
}

// kickDrain signals the drain goroutine to run one extra
// iteration right away. Non-blocking: if a kick is already
// pending the new one coalesces with it.
func (b *Bus) kickDrain() {
	if b.drainKick == nil {
		return
	}
	select {
	case b.drainKick <- struct{}{}:
	default:
	}
}

// Subscribe registers a sync handler. Sync handlers
// run on the publishing goroutine before Publish
// returns; reserve for cache invalidation and other
// in-process reactions.
func Subscribe[T Event](
	bus *Bus,
	handler Handler[T],
	opts ...SubscribeOption,
) {
	var zero T
	cfg := newSubConfig(opts)
	bus.add(zero.EventType(), subscription{
		name:     handlerName(handler),
		critical: cfg.critical,
		invoke: func(ctx context.Context, e Event) error {
			typed, ok := e.(T)
			if !ok {
				return fmt.Errorf(
					"events: type mismatch for %s: got %T",
					zero.EventType(), e,
				)
			}
			return handler(ctx, typed)
		},
	})
}

// SubscribeAsync registers an async handler. Each
// publish becomes one Asynq task per async handler;
// the event-dispatcher worker invokes the handler later.
func SubscribeAsync[T Event](
	bus *Bus,
	handler Handler[T],
	opts ...SubscribeOption,
) {
	var zero T
	cfg := newSubConfig(opts)
	if cfg.taskIDFn == nil {
		panic(fmt.Sprintf(
			"events: SubscribeAsync for %s requires "+
				"WithTaskIDFn for idempotency",
			zero.EventType(),
		))
	}
	bus.add(zero.EventType(), subscription{
		name:      handlerName(handler),
		isAsync:   true,
		asyncOpts: cfg.asyncOpts,
		taskIDFn:  cfg.taskIDFn,
		invoke: func(ctx context.Context, e Event) error {
			typed, ok := e.(T)
			if !ok {
				return fmt.Errorf(
					"events: type mismatch for %s: got %T",
					zero.EventType(), e,
				)
			}
			return handler(ctx, typed)
		},
	})
}

// Publish dispatches evt to sync and async subscribers.
//
// Behavior depends on whether ctx carries an active DB
// transaction (set by persistence.Manager.WithTx) AND
// whether the event has any async subscribers:
//
//  1. Inside a tx with async subscribers: the bus writes
//     the outbox row INSIDE the caller's transaction and
//     returns nil. Sync handlers and async enqueue are
//     deferred — they happen later when the drain
//     goroutine re-publishes the row in drain context,
//     well after the tx has committed. This is the
//     mandatory pattern for any event that follows a
//     domain mutation.
//
//  2. Inside a tx with NO async subscribers: the bus
//     runs sync handlers immediately on the request
//     goroutine. No outbox row is written. Sync handlers
//     observe pre-commit state — only safe when the
//     handler reads no DB row that the surrounding tx is
//     about to mutate.
//
//  3. Outside a tx: the bus writes the outbox row (in a
//     standalone connection), runs sync handlers, and
//     attempts async enqueue. On enqueue failure the
//     row remains pending and the drain picks it up.
//
//  4. Called from the drain itself (drainCtx is set):
//     no new outbox row is written; sync + async run
//     normally. The drain owns the outbox row's
//     MarkProcessed/MarkFailed bookkeeping.
//
// Returns the first critical sync error, if any. Panics
// in sync handlers are recovered and logged, never
// propagated.
func (b *Bus) Publish(ctx context.Context, evt Event) error {
	b.mu.RLock()
	subs := append(
		[]subscription(nil),
		b.subs[evt.EventType()]...,
	)
	b.mu.RUnlock()

	if len(subs) == 0 {
		return nil
	}

	metrics.EventsPublishedTotal.
		WithLabelValues(evt.EventType()).
		Inc()

	fromDrain := ctx.Value(drainKey{}) != nil

	var hasAsync bool
	for _, s := range subs {
		if s.isAsync {
			hasAsync = true
			break
		}
	}

	// Fast path 1: caller is the drain itself — run sync
	// and try async; drain owns outbox row bookkeeping.
	if fromDrain {
		return b.dispatch(ctx, evt, subs)
	}

	// Fast path 2: caller is inside a tx. Write the outbox
	// row in the caller's tx so the row shares commit fate
	// with the mutation. Sync (cache invalidation) and
	// async (notifications, lifecycle) dispatch happen
	// later from drainCtx — AFTER the tx commits.
	//
	// This applies to sync-only events too. Running cache
	// invalidation before commit would create a race: a
	// concurrent reader could repopulate the cache with
	// the pre-commit value, leaving stale data after the
	// tx commits. The 5 s drain interval means the cache
	// is fresh within seconds of commit.
	if persistence.HasTx(ctx) && b.outbox != nil {
		if _, err := b.outbox.Write(ctx, evt); err != nil {
			// Hard error: surface to caller so the tx
			// rolls back. Losing the event silently is
			// what this whole machinery exists to prevent.
			return fmt.Errorf(
				"events: outbox write in tx failed: %w", err,
			)
		}
		// Kick the drain so dispatch runs ASAP after
		// commit instead of waiting for the next tick.
		b.kickDrain()
		return nil
	}

	// Slow path: caller is OUTSIDE a tx. Write the outbox
	// row in a standalone connection (atomic write of one
	// row only — no surrounding mutation to share commit
	// fate with). Then run sync and try async.
	var (
		outboxID int64
		err      error
	)
	if hasAsync && b.outbox != nil {
		outboxID, err = b.outbox.Write(ctx, evt)
		if err != nil {
			b.logger.Warn(
				"outbox write failed; event may be lost on crash",
				"event", evt.EventType(),
				"error", err,
			)
		}
	}

	dispatchErr := b.dispatch(ctx, evt, subs)

	if outboxID > 0 {
		// Outbox row bookkeeping: mark processed only if
		// every async handler enqueued successfully.
		// dispatch returns ErrEnqueueFailed when any
		// async enqueue failed.
		if errors.Is(dispatchErr, ErrEnqueueFailed) {
			_ = b.outbox.MarkFailed(
				ctx, outboxID,
				"one or more enqueues failed",
			)
			b.kickDrain()
		} else {
			_ = b.outbox.MarkProcessed(ctx, outboxID)
		}
	}

	return dispatchErr
}

// dispatch runs sync handlers in registration order and
// attempts async enqueue for each async handler. Returns
// the first critical sync error, or ErrEnqueueFailed if
// any async enqueue failed. Sync errors from
// non-critical handlers are logged and otherwise ignored.
func (b *Bus) dispatch(
	ctx context.Context,
	evt Event,
	subs []subscription,
) error {
	var firstCritErr error
	allEnqueued := true

	for _, s := range subs {
		if s.isAsync {
			if b.asynq == nil {
				b.logger.Warn(
					"event async handler skipped: dispatcher not configured",
					"event",
					evt.EventType(),
					"handler",
					s.name,
				)
				allEnqueued = false
				continue
			}
			opts := s.asyncOpts
			if s.taskIDFn != nil {
				opts = append(
					append([]asynq.Option(nil), opts...),
					asynq.TaskID(s.taskIDFn(evt)),
				)
			}
			if err := b.asynq.Enqueue(
				ctx, evt, opts,
			); err != nil {
				b.logger.Error(
					"event async enqueue failed",
					"event", evt.EventType(),
					"handler", s.name,
					"error", err,
				)
				allEnqueued = false
				metrics.EventsEnqueueFailuresTotal.
					WithLabelValues(evt.EventType()).
					Inc()
			}
			continue
		}

		if err := b.runSync(ctx, evt, s); err != nil {
			b.logger.Warn(
				"event sync handler failed",
				"event", evt.EventType(),
				"handler", s.name,
				"error", err,
			)
			if s.critical && firstCritErr == nil {
				firstCritErr = err
			}
		}
	}

	if !allEnqueued {
		return ErrEnqueueFailed
	}
	return firstCritErr
}

// DrainOutbox polls the outbox table at the given
// interval and re-publishes unprocessed events. It also
// drains immediately whenever Publish kicks the
// drainKick channel — typically right after a publish
// inside a transaction commits, or after an async
// enqueue fails. Call this in a goroutine after Bus is
// wired; it blocks until ctx is canceled.
//
// Concurrent drainers across processes are safe:
// PendingBatch uses FOR UPDATE SKIP LOCKED, so each row
// is delivered at most once. Most replicas waste a
// cheap query per tick when no work is available.
func (b *Bus) DrainOutbox(
	ctx context.Context,
	interval time.Duration,
) {
	if b.outbox == nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.drainOnce(ctx)
		case <-b.drainKick:
			// Immediate-drain trigger. A small backoff
			// would let multiple kicks coalesce, but
			// PendingBatch is cheap and the kick channel
			// already buffers at most one extra wake-up.
			b.drainOnce(ctx)
		}
	}
}

func (b *Bus) drainOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			metrics.WorkerPanicsTotal.
				WithLabelValues("outbox_drain").
				Inc()
			b.logger.Error(
				"outbox drain: panic recovered",
				"panic", r,
				"stack_trace", string(debug.Stack()),
			)
		}
	}()

	// Circuit breaker: when the previous N drain passes all failed
	// (every row's Publish bounced, typically because Asynq/Redis is
	// unreachable), skip the body until the cooldown elapses. A
	// successful pass after the cooldown closes the breaker.
	if b.drainBreakerFailures >= drainBreakerThreshold {
		if time.Since(b.drainBreakerOpenedAt) < drainBreakerCooldown {
			return
		}
		b.logger.Info(
			"outbox drain: breaker cooldown elapsed; retrying",
			"consecutive_failures", b.drainBreakerFailures,
		)
	}

	rows, err := b.outbox.PendingBatch(ctx, 100)
	if err != nil {
		b.logger.Error(
			"outbox drain: fetch pending failed",
			"error", err,
		)
		return
	}
	if len(rows) == 0 {
		return
	}

	// Count publish failures so the breaker can decide whether the
	// drain pass was a total failure (every row bounced).
	var publishFailures, publishAttempts int

	drainCtx := context.WithValue(ctx, drainKey{}, true)
	for _, row := range rows {
		if row.RetryCount >= outboxMaxRetry {
			lastErr := ""
			if row.LastError != nil {
				lastErr = *row.LastError
			}
			if dlErr := b.outbox.DeadLetter(
				ctx, row, "retry_exhausted", lastErr,
			); dlErr != nil {
				b.logger.Error(
					"outbox drain: dead-letter failed",
					"id", row.ID, "error", dlErr,
				)
			}
			if mpErr := b.outbox.MarkProcessed(
				ctx, row.ID,
			); mpErr != nil {
				b.logger.Error(
					"outbox drain: mark processed failed",
					"id", row.ID, "error", mpErr,
				)
			}
			continue
		}

		factory := FactoryFor(row.EventType)
		if factory == nil {
			b.logger.Warn(
				"outbox drain: unknown event type",
				"type", row.EventType,
				"id", row.ID,
			)
			if dlErr := b.outbox.DeadLetter(
				ctx, row, "unknown_event_type", "",
			); dlErr != nil {
				b.logger.Error(
					"outbox drain: dead-letter failed",
					"id", row.ID, "error", dlErr,
				)
			}
			if mpErr := b.outbox.MarkProcessed(
				ctx, row.ID,
			); mpErr != nil {
				b.logger.Error(
					"outbox drain: mark processed failed",
					"id", row.ID, "error", mpErr,
				)
			}
			continue
		}

		evt := factory()
		if err := json.Unmarshal(row.Payload, evt); err != nil {
			b.logger.Error(
				"outbox drain: unmarshal failed",
				"type", row.EventType,
				"id", row.ID,
				"error", err,
			)
			if dlErr := b.outbox.DeadLetter(
				ctx, row, "unmarshal_failed", err.Error(),
			); dlErr != nil {
				b.logger.Error(
					"outbox drain: dead-letter failed",
					"id", row.ID, "error", dlErr,
				)
			}
			if mpErr := b.outbox.MarkProcessed(
				ctx, row.ID,
			); mpErr != nil {
				b.logger.Error(
					"outbox drain: mark processed failed",
					"id", row.ID, "error", mpErr,
				)
			}
			continue
		}

		publishAttempts++
		if err := b.Publish(drainCtx, evt); err != nil {
			publishFailures++
			if mfErr := b.outbox.MarkFailed(
				ctx, row.ID, err.Error(),
			); mfErr != nil {
				b.logger.Error(
					"outbox drain: mark failed failed",
					"id", row.ID, "error", mfErr,
				)
			}
		} else {
			if mpErr := b.outbox.MarkProcessed(
				ctx, row.ID,
			); mpErr != nil {
				b.logger.Error(
					"outbox drain: mark processed failed",
					"id", row.ID, "error", mpErr,
				)
			}
		}
	}

	// Update the circuit-breaker state. A total-failure pass
	// (every attempted publish bounced) increments the consecutive-
	// failure counter; any successful publish closes the breaker.
	switch {
	case publishAttempts == 0:
		// Nothing to publish — counter unchanged.
	case publishFailures == publishAttempts:
		b.drainBreakerFailures++
		if b.drainBreakerFailures == drainBreakerThreshold {
			b.drainBreakerOpenedAt = time.Now()
			b.logger.Warn(
				"outbox drain: breaker opened",
				"consecutive_failures", b.drainBreakerFailures,
				"cooldown", drainBreakerCooldown,
			)
		}
	default:
		if b.drainBreakerFailures >= drainBreakerThreshold {
			b.logger.Info(
				"outbox drain: breaker closed",
				"consecutive_failures_before_close",
				b.drainBreakerFailures,
			)
		}
		b.drainBreakerFailures = 0
	}
}

// InvokeAsync runs only the async subscriptions for an
// event - sync ones already ran on the publisher's
// goroutine. Called by Processor on the worker side.
func (b *Bus) InvokeAsync(
	ctx context.Context,
	evt Event,
) error {
	b.mu.RLock()
	subs := append(
		[]subscription(nil),
		b.subs[evt.EventType()]...,
	)
	b.mu.RUnlock()

	var firstErr error
	for _, s := range subs {
		if !s.isAsync {
			continue
		}
		err := b.runSync(ctx, evt, s)
		if err != nil {
			b.logger.Error(
				"event async handler failed",
				"event", evt.EventType(),
				"handler", s.name,
				"error", err,
			)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// HasSubscribers reports whether any handler is registered for
// eventType. Useful for short-circuiting expensive payload
// construction when no one is listening.
func (b *Bus) HasSubscribers(eventType string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.subs[eventType]) > 0
}

func (b *Bus) add(eventType string, sub subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs[eventType] = append(b.subs[eventType], sub)
}

func (b *Bus) runSync(
	ctx context.Context,
	evt Event,
	s subscription,
) (err error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Seconds()
		metrics.EventsHandlerDurationSeconds.
			WithLabelValues(evt.EventType(), s.name, "sync").
			Observe(elapsed)

		if r := recover(); r != nil {
			err = fmt.Errorf("events: handler panic: %v", r)
			b.logger.Error(
				"event handler panic recovered",
				"event", evt.EventType(),
				"handler", s.name,
				"panic", r,
			)
		}
	}()

	return s.invoke(ctx, evt)
}

func handlerName(h any) string {
	v := reflect.ValueOf(h)
	if !v.IsValid() || v.Kind() != reflect.Func {
		return "<unknown>"
	}
	if fn := runtime.FuncForPC(v.Pointer()); fn != nil {
		return fn.Name()
	}
	return "<unknown>"
}
