package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/isyll/go-grpc-starter/internal/models"

	"gorm.io/gorm"
)

type TokenRepository interface {
	Upsert(ctx context.Context, token *models.FCMToken) error
	ListByUserID(
		ctx context.Context, userID int64,
	) ([]*models.FCMToken, error)
	FindByUserIDAndDeviceID(
		ctx context.Context,
		userID int64,
		deviceID string,
	) (*models.FCMToken, error)
	DeleteByDeviceID(
		ctx context.Context, userID int64, deviceID string,
	) error
}

type PreferencesRepository interface {
	FindByUserID(
		ctx context.Context, userID int64,
	) (*models.NotificationPreferences, error)
	Upsert(
		ctx context.Context,
		prefs *models.NotificationPreferences,
	) error
	Create(
		ctx context.Context,
		prefs *models.NotificationPreferences,
	) error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Upsert(
	ctx context.Context, token *models.FCMToken,
) error {
	if err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND device_id = ?",
			token.UserID, token.DeviceID,
		).
		Assign(map[string]any{
			"token":       token.Token,
			"platform":    token.Platform,
			"app_version": token.AppVersion,
			"is_active":   true,
			"updated_at":  time.Now().UTC(),
		}).
		FirstOrCreate(token).Error; err != nil {
		panic(fmt.Errorf("upsert FCM token: %w", err))
	}
	return nil
}

func (r *tokenRepository) ListByUserID(
	ctx context.Context, userID int64,
) ([]*models.FCMToken, error) {
	var tokens []*models.FCMToken
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokens).Error; err != nil {
		panic(fmt.Errorf(
			"list FCM tokens for user %d: %w", userID, err,
		))
	}
	return tokens, nil
}

func (r *tokenRepository) FindByUserIDAndDeviceID(
	ctx context.Context,
	userID int64,
	deviceID string,
) (*models.FCMToken, error) {
	var token models.FCMToken
	err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND device_id = ?",
			userID, deviceID,
		).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		panic(fmt.Errorf("find FCM token: %w", err))
	}
	return &token, nil
}

func (r *tokenRepository) DeleteByDeviceID(
	ctx context.Context, userID int64, deviceID string,
) error {
	if err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND device_id = ?",
			userID, deviceID,
		).
		Delete(&models.FCMToken{}).Error; err != nil {
		panic(fmt.Errorf("delete FCM token: %w", err))
	}
	return nil
}

type preferencesRepository struct {
	db *gorm.DB
}

func NewPreferencesRepository(db *gorm.DB) PreferencesRepository {
	return &preferencesRepository{db: db}
}

func (r *preferencesRepository) FindByUserID(
	ctx context.Context, userID int64,
) (*models.NotificationPreferences, error) {
	var prefs models.NotificationPreferences
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&prefs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPrefNotFound
		}
		panic(fmt.Errorf(
			"find notification preferences: %w", err,
		))
	}
	return &prefs, nil
}

func (r *preferencesRepository) Create(
	ctx context.Context,
	prefs *models.NotificationPreferences,
) error {
	if err := r.db.WithContext(ctx).Create(prefs).Error; err != nil {
		panic(fmt.Errorf(
			"create notification preferences: %w", err,
		))
	}
	return nil
}

func (r *preferencesRepository) Upsert(
	ctx context.Context,
	prefs *models.NotificationPreferences,
) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", prefs.UserID).
		Assign(prefs).
		FirstOrCreate(prefs).Error; err != nil {
		panic(fmt.Errorf(
			"upsert notification preferences: %w", err,
		))
	}
	return nil
}
