// Package events implements the in-process event bus and the
// background Asynq fan-out used by App to react to cross-domain
// state changes without coupling domain services to one another.
//
// A domain service that mutates state publishes a typed event
// (past-tense fact, never a command). Subscribers fall into two
// categories:
//
//   - Synchronous cache invalidators registered with Subscribe;
//     these run when the drain re-publishes the outbox row (or
//     immediately, when Publish is called outside a tx), perform
//     no I/O beyond Redis tag invalidation, and their errors are
//     logged but never fail the request unless WithCritical is set.
//   - Asynchronous fan-out handlers registered with SubscribeAsync;
//     these enqueue the event to the events:dispatch Asynq queue,
//     a worker (cmd/worker/event_dispatcher) deserialises it via the
//     registered factory and re-enters the bus to run the handler.
//
// Transaction-aware publishing: events that follow a domain mutation
// MUST be published INSIDE the originating persistence.Manager.WithTx
// closure. The bus detects the tx on the context, writes the outbox
// row in the same transaction, and returns. Sync and async dispatch
// happen later when the drain re-publishes the row in drainCtx —
// strictly AFTER commit. This eliminates the
// mutation-committed-event-lost crash window and the
// invalidate-before-commit cache race in a single design.
//
// Async handlers MUST be idempotent: Asynq redelivers on Redis
// hiccups. The outbox writer offers at-least-once delivery for
// handlers whose effects must not be lost.
//
// Contract summary:
//
//   - Events are past-tense facts (user.registered), never
//     commands (user.register).
//   - Event types use pointer receivers so the async processor can
//     JSON-unmarshal into the factory's return value.
//   - Every new event type registers a factory in types.go's init().
//   - Sync handlers are reserved for cache invalidation. Anything
//     doing network I/O uses SubscribeAsync.
//   - Every async subscription supplies an idempotency key and a
//     uniqueness window.
//   - Cross-domain side effects belong on the bus, not in direct
//     dispatcher calls injected into domain services.
//   - Every Publish that follows a mutation runs inside
//     persistence.Manager.WithTx so the outbox row shares commit
//     fate with the mutation.
package events

import "context"

// Event is anything emitted on the bus. EventType must
// be a stable, snake_case routing key.
type Event interface {
	EventType() string
}

// Dedupable is the optional interface events implement to opt into
// outbox-level deduplication. When the bus writes an outbox row for
// a Dedupable event whose OutboxDedupeKey() returns a non-empty
// string, a partial unique index blocks a second pending row with
// the same (EventType, dedupe key) tuple. Events that return "" are
// treated as non-Dedupable.
type Dedupable interface {
	OutboxDedupeKey() string
}

// Handler receives an event of one concrete type. Sync
// handler errors are informational unless registered
// WithCritical(); async handler errors trigger Asynq
// retries.
type Handler[T Event] func(ctx context.Context, evt T) error
