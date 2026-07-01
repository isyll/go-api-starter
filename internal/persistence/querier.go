package persistence

import (
	"context"

	"gorm.io/gorm"
)

// Q returns the *gorm.DB bound to ctx by Manager.WithTx.
// Falls back to a fresh session on fallback when no transaction is
// active, preserving ctx for statement timeout propagation.
func Q(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

// HasTx reports whether ctx carries an active transaction
// previously installed by Manager.WithTx. Use it when a
// decision depends on tx participation rather than just
// the gorm.DB handle — e.g. the event bus picks a
// different publish strategy when ctx is transactional.
func HasTx(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(*gorm.DB)
	return ok
}
