package persistence

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type txKey struct{}

type Manager struct {
	db *gorm.DB
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

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
