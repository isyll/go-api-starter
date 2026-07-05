---
name: add-event
description: Add a domain event with a sync or async handler through the transactional outbox. Use when the user wants something to happen after a domain change (audit, cache invalidation, notification, projection).
---

# Add a domain event and handler

## Steps

1. Define the event in `internal/event/types_<domain>.go`: a struct with
   `CommonFields`, an `EventType() string` method returning a stable
   dotted name (`billing.invoice_paid`), and registration in the factory
   registry (see `internal/event/registry.go` and existing types files).
2. Publish it from the domain service with `bus.Publish(ctx, evt)` inside
   the same `store.WithTx` transaction as the domain write. That is what
   makes delivery transactional: the outbox row commits with the change.
3. Write the handler in `internal/event/handlers/`.
4. Subscribe in `internal/app/WireEventSubscriptions`
   (`internal/app/wire_events.go`):
   - `event.Subscribe` for fast, in-process side effects (cache
     invalidation). Use `event.WithCritical()` only if a failure must
     abort the publishing request.
   - `event.SubscribeAsync` for anything slow or retryable. Required:
     `event.WithTaskIDFn` returning a deterministic id (this is the
     idempotency key across redeliveries). Choose a queue with
     `event.WithQueue(event.QueueHigh|QueueNormal|QueueLow)`; add
     `event.WithUniqueWindow`, `event.WithMaxRetry`, or
     `event.WithTimeout` when the defaults (10 retries, 30s) do not fit.
5. Async handlers must be idempotent: they run at least once and can run
   twice after a crash. Upserts and dedupe keys, not blind inserts.
6. Test the handler, and extend `internal/event/bus_test.go` only when the
   bus contract itself changes.
7. Verify: `go build ./...`, `just test`.
