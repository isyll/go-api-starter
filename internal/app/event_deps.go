package app

import (
	"github.com/isyll/go-api-starter/internal/infra/cache"
	"github.com/isyll/go-api-starter/pkg/logger"

	"gorm.io/gorm"
)

// EventHandlerDeps is the set of dependencies the event handlers need.
// Handlers pull what they need from this struct so the subscription
// wiring stays stable as handlers are added.
type EventHandlerDeps struct {
	DB           *gorm.DB
	CacheManager *cache.CacheManager
	Logger       *logger.Logger
}
