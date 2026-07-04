package event

import "sync"

type Factory func() Event

var (
	registryMu sync.RWMutex
	registry   = map[string]Factory{}
)

func Register(f Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	zero := f()
	registry[zero.EventType()] = f
}

func FactoryFor(eventType string) Factory {
	registryMu.RLock()
	defer registryMu.RUnlock()

	return registry[eventType]
}
