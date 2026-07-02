package app

import (
	"github.com/isyll/go-api-starter/internal/infra/cache"
	"github.com/isyll/go-api-starter/pkg/logger"

	"gorm.io/gorm"
)

type EventHandlerDeps struct {
	DB           *gorm.DB
	CacheManager *cache.CacheManager
	Logger       *logger.Logger
}
