// Package suspension manages account suspensions used by moderation.
package suspension

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/isyll/go-grpc-starter/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetActiveByUserID(ctx context.Context, userID int64) (*models.AccountSuspension, error)
	Create(ctx context.Context, s *models.AccountSuspension) error
	DeactivateByUserID(ctx context.Context, userID int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetActiveByUserID(
	ctx context.Context, userID int64,
) (*models.AccountSuspension, error) {
	var sus models.AccountSuspension
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("is_permanent = true OR suspended_until > NOW()").
		Order("created_at DESC").
		First(&sus).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotSuspended
		}
		panic(fmt.Errorf("get active suspension: %w", err))
	}
	return &sus, nil
}

func (r *repository) Create(ctx context.Context, s *models.AccountSuspension) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("create suspension: %w", err)
	}
	return nil
}

func (r *repository) DeactivateByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&models.AccountSuspension{}).
		Where("user_id = ?", userID).
		Where("is_permanent = true OR suspended_until > NOW()").
		Updates(map[string]any{
			"is_permanent":    false,
			"suspended_until": time.Now().UTC(),
		}).Error
}
