// Package events implements the in-process event bus and outbox.
package events

import "context"

type Event interface {
	EventType() string
}

type Dedupable interface {
	OutboxDedupeKey() string
}

type Handler[T Event] func(ctx context.Context, evt T) error
