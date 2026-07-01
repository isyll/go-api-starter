package handlers

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/isyll/go-api-starter/internal/events"
	"github.com/isyll/go-api-starter/internal/models"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// AuthAttemptHandler persists AuthAttemptRecorded events as rows
// in auth.login_attempts on the main app_owner database. Running the
// write inside the event-dispatcher worker keeps the insert off
// the API request hot path while giving it outbox durability.
type AuthAttemptHandler struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewAuthAttemptHandler wires an AuthAttemptHandler backed by the
// given database connection and logger.
func NewAuthAttemptHandler(
	db *gorm.DB,
	logx *logger.Logger,
) *AuthAttemptHandler {
	return &AuthAttemptHandler{db: db, logger: logx}
}

// OnAuthAttemptRecorded persists the event payload as an
// auth.login_attempts row. On a DB error it logs the failure and
// returns the error so Asynq can retry with backoff.
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
