package app

import (
	"github.com/isyll/go-grpc-starter/internal/infra/cache"
	"github.com/isyll/go-grpc-starter/pkg/logger"

	"gorm.io/gorm"
)

type EventHandlerDeps struct {
	DB           *gorm.DB
	CacheManager *cache.CacheManager
	Logger       *logger.Logger
}
