// Package settings owns per-user preferences stored as a JSONB blob.
package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/isyll/go-api-starter/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID int64) (*models.Settings, error)
	Create(ctx context.Context, s *models.UserSettings) error
	Update(ctx context.Context, userID int64, s models.Settings) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (*models.Settings, error) {
	var row models.UserSettings
	err := r.db.WithContext(ctx).First(&row, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettingsNotFound
		}
		panic(fmt.Errorf("get settings: %w", err))
	}
	return &row.Settings, nil
}

func (r *repository) Create(ctx context.Context, s *models.UserSettings) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("create settings: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, userID int64, s models.Settings) error {
	return r.db.WithContext(ctx).Model(&models.UserSettings{}).
		Where("user_id = ?", userID).
		Update("settings", &s).Error
}
