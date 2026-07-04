package app

import (
	"fmt"
	"time"

	"github.com/isyll/go-grpc-starter/internal/event"
	"github.com/isyll/go-grpc-starter/internal/event/handlers"
)

func WireEventSubscriptions(bus *event.Bus, deps *EventHandlerDeps) {
	cacheInv := handlers.NewCacheInvalidator(deps.CacheManager, deps.Logger)
	event.Subscribe(bus, cacheInv.OnUserAccountDeleted)

	audit := handlers.NewAuditLogHandler(deps.Store, deps.Logger)
	event.SubscribeAsync(
		bus, audit.OnAuditLogWritten,
		event.WithQueue("high"),
		event.WithUniqueWindow(5*time.Minute),
		event.WithTaskIDFn(func(e event.Event) string {
			if evt, ok := e.(*event.AuditLogWritten); ok {
				return fmt.Sprintf("audit:%s:%s", evt.RequestID, evt.Action)
			}
			return ""
		}),
	)

	attempts := handlers.NewAuthAttemptHandler(deps.Store, deps.Logger)
	event.SubscribeAsync(
		bus, attempts.OnAuthAttemptRecorded,
		event.WithQueue("normal"),
		event.WithUniqueWindow(5*time.Minute),
		event.WithTaskIDFn(func(e event.Event) string {
			if evt, ok := e.(*event.AuthAttemptRecorded); ok {
				return fmt.Sprintf("auth:%s:%s", evt.RequestID, evt.Channel)
			}
			return ""
		}),
	)
}
