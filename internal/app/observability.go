package app

import (
	"context"
	"net/http"
	"os"

	"github.com/isyll/go-grpc-starter/internal/platform/obs"
)

func (a *App) startObservability() *http.Server {
	port := os.Getenv("METRICS_PORT")
	if port == "" {
		port = "9090"
	}
	return obs.StartServer(port, map[string]obs.Check{
		"database": func(ctx context.Context) error {
			return a.Infra.Store.Pool().Ping(ctx)
		},
		"cache": func(ctx context.Context) error {
			return a.Infra.Cache.Ping(ctx).Err()
		},
	}, a.Infra.Logger)
}
