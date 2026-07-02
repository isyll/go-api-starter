package grpc

import (
	"context"

	apiv1 "github.com/isyll/go-api-starter/internal/gen/api/v1"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type HealthServer struct {
	apiv1.UnimplementedHealthServiceServer
	db      *gorm.DB
	cache   *redis.Client
	version string
}

func NewHealthServer(db *gorm.DB, cache *redis.Client, version string) *HealthServer {
	return &HealthServer{db: db, cache: cache, version: version}
}

func (s *HealthServer) Check(_ context.Context, _ *emptypb.Empty) (*apiv1.HealthResponse, error) {
	return &apiv1.HealthResponse{Status: "ok", Version: s.version}, nil
}

func (s *HealthServer) Ready(ctx context.Context, _ *emptypb.Empty) (*apiv1.HealthResponse, error) {
	checks := map[string]string{"database": "ok", "cache": "ok"}
	status := "ok"

	if sqlDB, err := s.db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
		checks["database"] = "down"
		status = "degraded"
	}
	if err := s.cache.Ping(ctx).Err(); err != nil {
		checks["cache"] = "down"
		status = "degraded"
	}
	return &apiv1.HealthResponse{Status: status, Version: s.version, Checks: checks}, nil
}
