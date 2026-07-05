# Events and workers

Domain changes publish events on an in-process bus. Async handlers run
through a transactional outbox so no event is lost if the process
crashes after commit.

## Outbox

`Bus.Publish` writes the event to `events.events_outbox` in the same
transaction as the domain change. The event-dispatcher worker drains
pending rows (`FOR UPDATE SKIP LOCKED`, safe across replicas) and
re-publishes them, which enqueues Asynq tasks for async handlers.

## Workers

| Worker | Queue | Job |
| ------ | ----- | --- |
| `event_dispatcher` | outbox | drain outbox, run async handlers |
| `email_sender` | emails | send transactional email (Resend) |
| `push_notifications` | notifications | send FCM push |

## Adding a handler

1. Define the event type in `internal/event/types_*.go` and register it.
2. Publish it from a domain service with `bus.Publish`.
3. Subscribe a handler in `internal/app/wire_events.go`
   (`Subscribe` for sync, `SubscribeAsync` for outbox-backed async).
