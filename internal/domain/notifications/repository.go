package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/isyll/go-api-starter/internal/models"

	"gorm.io/gorm"
)

// TokenRepository persists notifications.fcm_tokens. Rows are
// keyed by (user_id, device_id) and the unique constraint is
// honored by Upsert.
type TokenRepository interface {
	// Upsert inserts or updates the FCM token row for the
	// (user, device) pair.
	Upsert(ctx context.Context, token *models.FCMToken) error
	// ListByUserID returns every active token for userID.
	ListByUserID(
		ctx context.Context, userID int64,
	) ([]*models.FCMToken, error)
	// FindByUserIDAndDeviceID returns the token row for the
	// exact (user, device) pair or ErrTokenNotFound.
	FindByUserIDAndDeviceID(
		ctx context.Context,
		userID int64,
		deviceID string,
	) (*models.FCMToken, error)
	// DeleteByDeviceID removes the (user, device) token row.
	DeleteByDeviceID(
		ctx context.Context, userID int64, deviceID string,
	) error
}

// PreferencesRepository persists
// notifications.user_notification_preferences. One row per
// user; the migration seeds defaults on user creation.
type PreferencesRepository interface {
	// FindByUserID returns the preferences row for userID, or
	// ErrPrefNotFound if no row exists.
	FindByUserID(
		ctx context.Context, userID int64,
	) (*models.NotificationPreferences, error)
	// Upsert inserts or replaces the preferences row.
	Upsert(
		ctx context.Context,
		prefs *models.NotificationPreferences,
	) error
	// Create inserts the initial preferences row for a new
	// user.
	Create(
		ctx context.Context,
		prefs *models.NotificationPreferences,
	) error
}

// tokenRepository implements TokenRepository against
// notifications.fcm_tokens.
type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository constructs the FCM-token repository.
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
