package persistence

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type txKey struct{}

// Manager owns the application's *gorm.DB and exposes WithTx to run
// a closure inside a PostgreSQL transaction that is bound to ctx.
type Manager struct {
	db *gorm.DB
}

// NewManager creates a Manager wrapping db.
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// WithTx runs fn inside a database transaction. The transaction is
// committed when fn returns nil and rolled back on any error.
// opts may be used to set the isolation level or read-only flag.
func (m *Manager) WithTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
	opts ...*sql.TxOptions,
) error {
	var to *sql.TxOptions
	if len(opts) > 0 {
		to = opts[0]
	}
	return m.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			return fn(
				context.WithValue(ctx, txKey{}, tx),
			)
		},
		to,
	)
}
