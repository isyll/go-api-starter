package grpcsvc

import (
	"context"

	healthv1 "github.com/isyll/go-grpc-starter/gen/health/v1"
	"github.com/isyll/go-grpc-starter/internal/store"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HealthServer struct {
	healthv1.UnimplementedHealthServiceServer
	store   *store.Store
	cache   *redis.Client
	version string
}

func NewHealthServer(store *store.Store, cache *redis.Client, version string) *HealthServer {
	return &HealthServer{store: store, cache: cache, version: version}
}

func (s *HealthServer) Check(_ context.Context, _ *emptypb.Empty) (*healthv1.HealthResponse, error) {
	return &healthv1.HealthResponse{Status: "ok", Version: s.version}, nil
}

func (s *HealthServer) Ready(ctx context.Context, _ *emptypb.Empty) (*healthv1.HealthResponse, error) {
	checks := map[string]string{"database": "ok", "cache": "ok"}
	status := "ok"

	if err := s.store.Pool().Ping(ctx); err != nil {
		checks["database"] = "down"
		status = "degraded"
	}
	if err := s.cache.Ping(ctx).Err(); err != nil {
		checks["cache"] = "down"
		status = "degraded"
	}
	return &healthv1.HealthResponse{Status: status, Version: s.version, Checks: checks}, nil
}
