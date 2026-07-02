package persistence

import (
	"context"

	"gorm.io/gorm"
)

func Q(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

func HasTx(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(*gorm.DB)
	return ok
}
