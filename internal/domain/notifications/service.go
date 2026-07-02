// Package notifications manages push tokens and delivery preferences.
package notifications

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/messaging"

	"github.com/isyll/go-api-starter/internal/models"
	"github.com/isyll/go-api-starter/pkg/logger"
)

type RegisterTokenInput struct {
	DeviceID   string
	Token      string
	Platform   models.NotificationPlatform
	AppVersion string
}

type PreferencesUpdate struct {
	Push              *bool
	Email             *bool
	Marketing         *bool
	QuietHoursEnabled *bool
	QuietHoursStart   *string
	QuietHoursEnd     *string
	Timezone          *string
}

type Service struct {
	tokens    TokenRepository
	prefs     PreferencesRepository
	fcmClient *messaging.Client
	logger    *logger.Logger
}

func NewService(
	tokens TokenRepository,
	prefs PreferencesRepository,
	fcmClient *messaging.Client,
	logx *logger.Logger,
) *Service {
	return &Service{tokens: tokens, prefs: prefs, fcmClient: fcmClient, logger: logx}
}

func (s *Service) RegisterToken(ctx context.Context, userID int64, in RegisterTokenInput) error {
	return s.tokens.Upsert(ctx, &models.FCMToken{
		UserID:     userID,
		DeviceID:   in.DeviceID,
		Token:      in.Token,
		Platform:   in.Platform,
		AppVersion: in.AppVersion,
		IsActive:   true,
	})
}

func (s *Service) ListTokens(ctx context.Context, userID int64) ([]*models.FCMToken, error) {
	return s.tokens.ListByUserID(ctx, userID)
}

func (s *Service) DeleteToken(ctx context.Context, userID int64, deviceID string) error {
	return s.tokens.DeleteByDeviceID(ctx, userID, deviceID)
}

func (s *Service) GetPreferences(
	ctx context.Context, userID int64,
) (*models.NotificationPreferences, error) {
	prefs, err := s.prefs.FindByUserID(ctx, userID)
	if err == nil {
		return prefs, nil
	}
	if !errors.Is(err, ErrPrefNotFound) {
		return nil, err
	}
	defaults := &models.NotificationPreferences{
		UserID: userID, Push: true, Email: true, Marketing: false, Timezone: "UTC",
	}
	if cErr := s.prefs.Create(ctx, defaults); cErr != nil {
		return nil, cErr
	}
	return defaults, nil
}

func (s *Service) UpdatePreferences(
	ctx context.Context, userID int64, upd PreferencesUpdate,
) (*models.NotificationPreferences, error) {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}
	if upd.Push != nil {
		prefs.Push = *upd.Push
	}
	if upd.Email != nil {
		prefs.Email = *upd.Email
	}
	if upd.Marketing != nil {
		prefs.Marketing = *upd.Marketing
	}
	if upd.QuietHoursEnabled != nil {
		prefs.QuietHoursEnabled = *upd.QuietHoursEnabled
	}
	if upd.QuietHoursStart != nil {
		prefs.QuietHoursStart = upd.QuietHoursStart
	}
	if upd.QuietHoursEnd != nil {
		prefs.QuietHoursEnd = upd.QuietHoursEnd
	}
	if upd.Timezone != nil {
		prefs.Timezone = *upd.Timezone
	}
	if err := s.prefs.Upsert(ctx, prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}
