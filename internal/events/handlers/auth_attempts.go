package handlers

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/isyll/go-grpc-starter/internal/events"
	"github.com/isyll/go-grpc-starter/internal/models"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

type AuthAttemptHandler struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewAuthAttemptHandler(
	db *gorm.DB,
	logx *logger.Logger,
) *AuthAttemptHandler {
	return &AuthAttemptHandler{db: db, logger: logx}
}

func (h *AuthAttemptHandler) OnAuthAttemptRecorded(
	ctx context.Context,
	evt *events.AuthAttemptRecorded,
) error {
	occurredAt := evt.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	row := models.LoginAttempt{
		Email:     evt.Email,
		UserID:    evt.UserID,
		Channel:   evt.Channel,
		Outcome:   evt.Outcome,
		Remaining: evt.Remaining,
		CreatedAt: occurredAt,
	}

	if evt.IPAddress != "" {
		row.IPAddress = &evt.IPAddress
	}
	if evt.UserAgent != "" {
		row.UserAgent = &evt.UserAgent
	}
	if evt.DeviceID != "" {
		row.DeviceID = &evt.DeviceID
	}
	if evt.RequestID != "" {
		row.RequestID = &evt.RequestID
	}

	if err := h.db.WithContext(ctx).Create(&row).Error; err != nil {
		h.logger.Error(
			"auth attempt handler: failed to persist login attempt row",
			"error", err,
			"email", evt.Email,
			"channel", evt.Channel,
			"outcome", evt.Outcome,
			"request_id", evt.RequestID,
		)
		return err
	}

	return nil
}
