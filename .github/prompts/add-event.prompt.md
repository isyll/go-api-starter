---
mode: agent
description: Add a domain event with a handler through the transactional outbox
---

Add the requested event end to end:

1. Define the event struct in `internal/event/types_<domain>.go` with
   `CommonFields`, a stable `EventType()` name, and factory registration.
2. Publish with `bus.Publish(ctx, evt)` inside the same `store.WithTx`
   transaction as the domain write.
3. Handler in `internal/event/handlers/`; subscribe in
   `internal/app/wire_events.go`. Use `event.Subscribe` for fast sync
   side effects, `event.SubscribeAsync` with `event.WithTaskIDFn`
   (deterministic id) and a namespaced queue
   (`event.QueueHigh|QueueNormal|QueueLow`) for anything retryable.
4. Async handlers must be idempotent (at-least-once delivery).
5. Verify: `go build ./...`, `just test`.
