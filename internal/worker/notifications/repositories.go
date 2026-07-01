package notifications

import (
	"context"
	"time"

	"github.com/isyll/go-api-starter/internal/models"

	"gorm.io/gorm"
)

// Local interfaces matching the domain/notifications
// repository contracts. Only the methods used by the
// async processor are included. Go structural typing
// ensures compatibility with the domain implementations.

type FCMTokenRepository interface {
	FindActiveByUserID(
		ctx context.Context,
		userID int64,
	) ([]*models.FCMToken, error)
	DeactivateByID(ctx context.Context, id int64) error
	UpdateLastUsed(ctx context.Context, id int64) error
}

type NotificationPreferencesRepository interface {
	FindByUserID(
		ctx context.Context,
		userID int64,
	) (*models.NotificationPreferences, error)
}

type NotificationTemplateRepository interface {
	FindByEventType(
		ctx context.Context,
		eventType string,
	) (*models.NotificationTemplate, error)
}

type NotificationLogRepository interface {
	Create(
		ctx context.Context,
		log *models.NotificationLog,
	) error
}

type fcmTokenRepository struct {
	db *gorm.DB
}

func NewFCMTokenRepository(db *gorm.DB) FCMTokenRepository {
	return &fcmTokenRepository{db: db}
}

func (r *fcmTokenRepository) FindActiveByUserID(
	ctx context.Context,
	userID int64,
) ([]*models.FCMToken, error) {
	var tokens []*models.FCMToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true", userID).
		Find(&tokens).Error
	return tokens, err
}

func (r *fcmTokenRepository) DeactivateByID(
	ctx context.Context,
	id int64,
) error {
	return r.db.WithContext(ctx).
		Model(&models.FCMToken{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *fcmTokenRepository) UpdateLastUsed(
	ctx context.Context,
	id int64,
) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.FCMToken{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}

type notifPreferencesRepository struct {
	db *gorm.DB
}

func NewNotificationPreferencesRepository(
	db *gorm.DB,
) NotificationPreferencesRepository {
	return &notifPreferencesRepository{db: db}
}

func (r *notifPreferencesRepository) FindByUserID(
	ctx context.Context,
	userID int64,
) (*models.NotificationPreferences, error) {
	var prefs models.NotificationPreferences
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&prefs).Error
	if err != nil {
		return nil, err
	}
	return &prefs, nil
}

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) NotificationTemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) FindByEventType(
	ctx context.Context,
	eventType string,
) (*models.NotificationTemplate, error) {
	var tmpl models.NotificationTemplate
	err := r.db.WithContext(ctx).
		Preload("Translations").
		Where("event_type = ?", eventType).
		First(&tmpl).Error
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) NotificationLogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) Create(
	ctx context.Context,
	log *models.NotificationLog,
) error {
	return r.db.WithContext(ctx).Create(log).Error
}
