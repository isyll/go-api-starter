package handlers

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/isyll/go-api-starter/internal/events"
	"github.com/isyll/go-api-starter/internal/metrics"
	"github.com/isyll/go-api-starter/internal/models"
	"github.com/isyll/go-api-starter/pkg/logger"
)

type AuditLogHandler struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewAuditLogHandler(db *gorm.DB, logx *logger.Logger) *AuditLogHandler {
	return &AuditLogHandler{db: db, logger: logx}
}

func (h *AuditLogHandler) OnAuditLogWritten(
	ctx context.Context,
	evt *events.AuditLogWritten,
) error {
	occurredAt := evt.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	status := evt.Status
	if status == "" {
		status = "success"
	}

	row := models.AuditLog{
		AdminID:    evt.AdminID,
		Action:     evt.Action,
		Resource:   evt.Resource,
		ResourceID: evt.ResourceID,
		Details:    evt.Details,
		Status:     status,
		IPAddress:  evt.IPAddress,
		UserAgent:  evt.UserAgent,
		RequestID:  evt.RequestID,
		CreatedAt:  occurredAt,
	}

	if err := h.db.WithContext(ctx).Create(&row).Error; err != nil {
		metrics.AuditLogWriteFailuresTotal.WithLabelValues("db_write").Inc()
		h.logger.Error(
			"audit log handler: failed to write audit row",
			"error", err, "action", evt.Action,
			"admin_id", evt.AdminID, "request_id", evt.RequestID,
		)
		return err
	}
	return nil
}
