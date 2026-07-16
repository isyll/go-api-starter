// Package obs serves metrics, liveness, readiness, and opt-in pprof.
package obs

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/isyll/go-grpc-starter/pkg/logger"
)

// Check probes one dependency for readiness.
type Check func(ctx context.Context) error

// StartServer starts the metrics/health server; PPROF_ENABLED also mounts pprof (keep private).
func StartServer(
	port string,
	checks map[string]Check,
	logx *logger.Logger,
) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		for name, check := range checks {
			if err := check(ctx); err != nil {
				logx.Warn("readiness check failed", "check", name, "error", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(name + ": unavailable"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	if os.Getenv("PPROF_ENABLED") == "true" {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		logx.Info("pprof enabled on observability port")
	}

	srv := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Error("observability server failed", "error", err)
		}
	}()
	logx.Info("observability server started", "port", port)
	return srv
}
