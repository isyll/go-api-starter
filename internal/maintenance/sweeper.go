// Package maintenance runs periodic retention sweeps on high-churn tables.
package maintenance

import (
	"context"
	"time"

	"github.com/isyll/go-grpc-starter/gen/db"
	"github.com/isyll/go-grpc-starter/internal/metrics"
	"github.com/isyll/go-grpc-starter/internal/store"
	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

type Sweeper struct {
	store  *store.Store
	cfg    *config.MaintenanceConfig
	logger *logger.Logger
}

func NewSweeper(
	st *store.Store,
	cfg *config.MaintenanceConfig,
	logx *logger.Logger,
) *Sweeper {
	return &Sweeper{store: st, cfg: cfg, logger: logx}
}

// Run sweeps once at startup, then every interval tick.
func (s *Sweeper) Run(ctx context.Context) {
	s.sweep(ctx)

	ticker := time.NewTicker(s.cfg.Retention.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sweep(ctx)
		}
	}
}

func (s *Sweeper) sweep(ctx context.Context) {
	now := time.Now().UTC()
	r := s.cfg.Retention

	s.run(ctx, "refresh_tokens", func(ctx context.Context, q *db.Queries) (int64, error) {
		return q.DeleteStaleRefreshTokens(ctx, store.TS(now.Add(-r.RefreshTokens)))
	})
	s.run(ctx, "processed_outbox", func(ctx context.Context, q *db.Queries) (int64, error) {
		return q.DeleteProcessedOutboxBefore(ctx, store.TS(now.Add(-r.ProcessedOutbox)))
	})
	s.run(ctx, "login_attempts", func(ctx context.Context, q *db.Queries) (int64, error) {
		return q.DeleteLoginAttemptsBefore(ctx, store.TS(now.Add(-r.LoginAttempts)))
	})
}

func (s *Sweeper) run(
	ctx context.Context,
	job string,
	del func(ctx context.Context, q *db.Queries) (int64, error),
) {
	var deleted int64
	err := s.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		var err error
		deleted, err = del(ctx, q)
		return err
	})
	if err != nil {
		s.logger.Error("maintenance sweep failed", "job", job, "error", err)
		return
	}
	metrics.MaintenanceRowsDeletedTotal.WithLabelValues(job).Add(float64(deleted))
	if deleted > 0 {
		s.logger.Info("maintenance sweep", "job", job, "rows_deleted", deleted)
	}
}
