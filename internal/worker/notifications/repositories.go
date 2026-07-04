package notifications

import (
	"context"
	"encoding/json"

	"github.com/isyll/go-grpc-starter/gen/db"
	domnotif "github.com/isyll/go-grpc-starter/internal/domain/notifications"
	"github.com/isyll/go-grpc-starter/internal/store"
)

type FCMTokenRepository interface {
	FindActiveByUserID(
		ctx context.Context,
		userID int64,
	) ([]*domnotif.FCMToken, error)
	DeactivateByID(ctx context.Context, id int64) error
	UpdateLastUsed(ctx context.Context, id int64) error
}

type NotificationPreferencesRepository interface {
	FindByUserID(
		ctx context.Context,
		userID int64,
	) (*domnotif.NotificationPreferences, error)
}

type NotificationTemplateRepository interface {
	FindByEventType(
		ctx context.Context,
		eventType string,
	) (*domnotif.NotificationTemplate, error)
}

type NotificationLogRepository interface {
	Create(
		ctx context.Context,
		log *domnotif.NotificationLog,
	) error
}

func toFCMToken(r db.AuthFcmToken) *domnotif.FCMToken {
	return &domnotif.FCMToken{
		ID:         r.ID,
		UserID:     r.UserID,
		DeviceID:   r.DeviceID,
		Token:      r.Token,
		Platform:   domnotif.NotificationPlatform(r.Platform),
		AppVersion: store.Str(r.AppVersion),
		IsActive:   r.IsActive,
		LastUsedAt: store.TimePtr(r.LastUsedAt),
		CreatedAt:  store.Time(r.CreatedAt),
		UpdatedAt:  store.Time(r.UpdatedAt),
	}
}

func toPreferences(r db.NotificationsNotificationPreference) *domnotif.NotificationPreferences {
	return &domnotif.NotificationPreferences{
		UserID:            r.UserID,
		Push:              r.Push,
		Email:             r.Email,
		Marketing:         r.Marketing,
		QuietHoursEnabled: r.QuietHoursEnabled,
		QuietHoursStart:   store.TimeOfDayStr(r.QuietHoursStart),
		QuietHoursEnd:     store.TimeOfDayStr(r.QuietHoursEnd),
		Timezone:          r.Timezone,
		CreatedAt:         store.Time(r.CreatedAt),
		UpdatedAt:         store.Time(r.UpdatedAt),
	}
}

func toTemplate(r db.NotificationsNotificationTemplate) *domnotif.NotificationTemplate {
	return &domnotif.NotificationTemplate{
		ID:               int(r.ID),
		EventType:        r.EventType,
		Icon:             r.Icon,
		Sound:            r.Sound,
		Priority:         r.Priority,
		AndroidChannelID: r.AndroidChannelID,
		Action:           r.Action,
		CreatedAt:        store.Time(r.CreatedAt),
		UpdatedAt:        store.Time(r.UpdatedAt),
	}
}

func toTemplateTranslation(
	r db.NotificationsNotificationTemplateTranslation,
) *domnotif.NotificationTemplateTranslation {
	return &domnotif.NotificationTemplateTranslation{
		ID:         int(r.ID),
		TemplateID: int(r.TemplateID),
		Language:   r.Language,
		Title:      r.Title,
		Body:       r.Body,
		CreatedAt:  store.Time(r.CreatedAt),
		UpdatedAt:  store.Time(r.UpdatedAt),
	}
}

type fcmTokenRepository struct {
	store *store.Store
}

func NewFCMTokenRepository(s *store.Store) FCMTokenRepository {
	return &fcmTokenRepository{store: s}
}

func (r *fcmTokenRepository) FindActiveByUserID(
	ctx context.Context,
	userID int64,
) ([]*domnotif.FCMToken, error) {
	var tokens []*domnotif.FCMToken
	err := r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		rows, err := q.ListActiveFCMTokensByUserID(ctx, userID)
		if err != nil {
			return err
		}
		tokens = make([]*domnotif.FCMToken, len(rows))
		for i, row := range rows {
			tokens[i] = toFCMToken(row)
		}
		return nil
	})
	return tokens, err
}

func (r *fcmTokenRepository) DeactivateByID(
	ctx context.Context,
	id int64,
) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.DeactivateFCMToken(ctx, id)
	})
}

func (r *fcmTokenRepository) UpdateLastUsed(
	ctx context.Context,
	id int64,
) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.TouchFCMTokenLastUsed(ctx, id)
	})
}

type notifPreferencesRepository struct {
	store *store.Store
}

func NewNotificationPreferencesRepository(
	s *store.Store,
) NotificationPreferencesRepository {
	return &notifPreferencesRepository{store: s}
}

func (r *notifPreferencesRepository) FindByUserID(
	ctx context.Context,
	userID int64,
) (*domnotif.NotificationPreferences, error) {
	var out *domnotif.NotificationPreferences
	err := r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		row, err := q.GetNotificationPreferences(ctx, userID)
		if err != nil {
			return err
		}
		out = toPreferences(row)
		return nil
	})
	return out, err
}

type templateRepository struct {
	store *store.Store
}

func NewTemplateRepository(s *store.Store) NotificationTemplateRepository {
	return &templateRepository{store: s}
}

func (r *templateRepository) FindByEventType(
	ctx context.Context,
	eventType string,
) (*domnotif.NotificationTemplate, error) {
	var out *domnotif.NotificationTemplate
	err := r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		row, err := q.GetNotificationTemplateByEventType(ctx, eventType)
		if err != nil {
			return err
		}
		tmpl := toTemplate(row)
		translations, err := q.ListTemplateTranslations(ctx, row.ID)
		if err != nil {
			return err
		}
		tmpl.Translations = make([]*domnotif.NotificationTemplateTranslation, len(translations))
		for i, tr := range translations {
			tmpl.Translations[i] = toTemplateTranslation(tr)
		}
		out = tmpl
		return nil
	})
	return out, err
}

type logRepository struct {
	store *store.Store
}

func NewLogRepository(s *store.Store) NotificationLogRepository {
	return &logRepository{store: s}
}

func (r *logRepository) Create(
	ctx context.Context,
	log *domnotif.NotificationLog,
) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		var payload []byte
		if log.Payload != nil {
			var err error
			payload, err = json.Marshal(log.Payload)
			if err != nil {
				return err
			}
		}
		_, err := q.CreateNotificationLog(ctx, db.CreateNotificationLogParams{
			UserID:       log.UserID,
			EventType:    log.EventType,
			EventID:      log.EventID,
			FcmTokenID:   log.FCMTokenID,
			Status:       db.NotificationsNotificationStatus(log.Status),
			ErrorCode:    log.ErrorCode,
			ErrorMessage: log.ErrorMessage,
			Payload:      payload,
		})
		return err
	})
}
