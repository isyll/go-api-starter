// Package store is the pgx + sqlc layer; every op runs in an RLS-scoped tx.
package store

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/isyll/go-grpc-starter/gen/db"
	"github.com/isyll/go-grpc-starter/internal/persistence"
	"github.com/isyll/go-grpc-starter/internal/reqctx"
)

type Store struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: db.New(pool)}
}

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

func InTx(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(*db.Queries)
	return ok
}

type (
	txKey    struct{}
	txRawKey struct{}
)

// Run runs fn on the ambient tx, else a fresh RLS-scoped tx.
func (s *Store) Run(
	ctx context.Context,
	fn func(ctx context.Context, q *db.Queries) error,
) error {
	if q, ok := ctx.Value(txKey{}).(*db.Queries); ok {
		persistence.IncrQueryCounter(ctx)
		return fn(ctx, q)
	}
	return s.WithTx(ctx, func(ctx context.Context) error {
		persistence.IncrQueryCounter(ctx)
		return fn(ctx, ctx.Value(txKey{}).(*db.Queries))
	})
}

// WithTx runs fn in one tx; composed writes commit atomically.
func (s *Store) WithTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	if _, ok := ctx.Value(txKey{}).(*db.Queries); ok {
		return fn(ctx) // already in a tx; join it
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := applyRLS(ctx, tx); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, txKey{}, s.q.WithTx(tx))
	ctx = context.WithValue(ctx, txRawKey{}, tx)
	if err := fn(ctx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SetChangeReason sets a GUC audit triggers record; no-op outside a tx.
func SetChangeReason(ctx context.Context, reason string) error {
	tx, ok := ctx.Value(txRawKey{}).(pgx.Tx)
	if !ok {
		return nil
	}
	_, err := tx.Exec(ctx, "SELECT set_config('app.change_reason', $1, true)", reason)
	return err
}

// applyRLS sets the per-request RLS GUCs the schema reads.
func applyRLS(ctx context.Context, tx pgx.Tx) error {
	s := reqctx.SubjectFrom(ctx)
	userID := "0"
	role := string(reqctx.RoleAnonymous)
	switch {
	case s.IsAdmin:
		role = string(reqctx.RoleAdmin)
	case s.UserID > 0:
		userID = strconv.FormatInt(s.UserID, 10)
		if s.Role != "" {
			role = string(s.Role)
		}
	}
	_, err := tx.Exec(
		ctx,
		`SELECT set_config('app.current_user_id', $1, true),
		        set_config('app.current_role', $2, true)`,
		userID, role,
	)
	return err
}
