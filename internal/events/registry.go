package events

import "sync"

// Factory builds a zero-value pointer the async
// processor unmarshals into.
type Factory func() Event

var (
	registryMu sync.RWMutex
	registry   = map[string]Factory{}
)

// Register makes an event type discoverable to the
// async processor. Call from init() in the file that
// declares the type.
func Register(f Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	zero := f()
	registry[zero.EventType()] = f
}

// FactoryFor returns the registered Factory for eventType, or nil
// when no type has registered. The async processor calls this to
// turn the queued envelope back into a concrete Event value.
func FactoryFor(eventType string) Factory {
	registryMu.RLock()
	defer registryMu.RUnlock()

	return registry[eventType]
}
