package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/isyll/go-api-starter/internal/metrics"
	"github.com/isyll/go-api-starter/internal/persistence"
	"github.com/isyll/go-api-starter/pkg/logger"
)

var ErrOutboxDuplicate = errors.New(
	"outbox: duplicate pending row with same (event_type, dedupe_key)",
)

type OutboxEvent struct {
	ID              int64           `gorm:"primaryKey"`
	EventType       string          `gorm:"not null"`
	Payload         json.RawMessage `gorm:"type:jsonb;not null"`
	DedupeKey       *string
	CreatedAt       time.Time `gorm:"not null;default:now()"`
	ProcessedAt     *time.Time
	RetryCount      int `gorm:"not null;default:0"`
	LastError       *string
	LastAttemptedAt *time.Time
}

func (OutboxEvent) TableName() string { return "events.outbox" }

type OutboxDeadLetterEvent struct {
	ID            int64           `gorm:"primaryKey"`
	SourceID      int64           `gorm:"not null"`
	EventType     string          `gorm:"not null"`
	Payload       json.RawMessage `gorm:"type:jsonb;not null"`
	FailureReason string          `gorm:"not null"`
	LastError     *string
	CreatedAt     time.Time `gorm:"not null;default:now()"`
	FailedAt      time.Time `gorm:"not null;default:now()"`
}

func (OutboxDeadLetterEvent) TableName() string {
	return "events.outbox_dead_letter"
}

type OutboxRepository struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewOutboxRepository(
	db *gorm.DB,
	logx *logger.Logger,
) *OutboxRepository {
	return &OutboxRepository{db: db, logger: logx}
}

func (r *OutboxRepository) Write(
	ctx context.Context,
	evt Event,
) (int64, error) {
	payload, err := json.Marshal(evt)
	if err != nil {
		return 0, fmt.Errorf("outbox: marshal: %w", err)
	}
	row := &OutboxEvent{
		EventType: evt.EventType(),
		Payload:   payload,
	}
	if d, ok := evt.(Dedupable); ok {
		if k := strings.TrimSpace(d.OutboxDedupeKey()); k != "" {
			row.DedupeKey = &k
		}
	}
	if err := persistence.Q(ctx, r.db).
		Create(row).Error; err != nil {
		if isOutboxDedupeViolation(err) {
			return 0, ErrOutboxDuplicate
		}
		return 0, fmt.Errorf("outbox: insert: %w", err)
	}
	return row.ID, nil
}

func isOutboxDedupeViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" &&
		strings.Contains(
			pgErr.ConstraintName, "outbox_pending_dedupe",
		)
}

func (r *OutboxRepository) Publish(
	ctx context.Context,
	evt Event,
) error {
	if _, err := r.Write(ctx, evt); err != nil {
		if errors.Is(err, ErrOutboxDuplicate) {
			r.logger.Debug(
				"outbox publish deduplicated",
				"event", evt.EventType(),
			)
			return nil
		}
		r.logger.Error(
			"outbox publish failed",
			"event", evt.EventType(),
			"error", err,
		)
		return err
	}
	return nil
}

func (r *OutboxRepository) MarkProcessed(
	ctx context.Context,
	id int64,
) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&OutboxEvent{}).
		Where("id = ?", id).
		Update("processed_at", now).Error; err != nil {
		metrics.OutboxMarkFailuresTotal.
			WithLabelValues("processed").Inc()
		r.logger.Warn(
			"outbox: mark processed failed",
			"id", id, "error", err,
		)
		return fmt.Errorf(
			"outbox: mark processed %d: %w", id, err,
		)
	}
	return nil
}

func (r *OutboxRepository) MarkFailed(
	ctx context.Context,
	id int64,
	errMsg string,
) error {
	if err := r.db.WithContext(ctx).
		Model(&OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"retry_count":       gorm.Expr("retry_count + 1"),
			"last_error":        errMsg,
			"last_attempted_at": time.Now(),
		}).Error; err != nil {
		metrics.OutboxMarkFailuresTotal.
			WithLabelValues("failed").Inc()
		r.logger.Warn(
			"outbox: mark failed failed",
			"id", id, "error", err,
		)
		return fmt.Errorf(
			"outbox: mark failed %d: %w", id, err,
		)
	}
	return nil
}

func (r *OutboxRepository) DeadLetter(
	ctx context.Context,
	row *OutboxEvent,
	reason string,
	lastErr string,
) error {
	var lastErrPtr *string
	if lastErr != "" {
		lastErrPtr = &lastErr
	}
	dl := &OutboxDeadLetterEvent{
		SourceID:      row.ID,
		EventType:     row.EventType,
		Payload:       row.Payload,
		FailureReason: reason,
		LastError:     lastErrPtr,
	}
	if err := r.db.WithContext(ctx).Create(dl).Error; err != nil {
		return fmt.Errorf(
			"outbox: dead-letter %d: %w", row.ID, err,
		)
	}
	metrics.OutboxDeadLetterTotal.WithLabelValues(reason).Inc()
	return nil
}

func (r *OutboxRepository) PendingBatch(
	ctx context.Context,
	limit int,
) ([]*OutboxEvent, error) {
	var rows []*OutboxEvent
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "SKIP LOCKED",
		}).
		Where(`
			processed_at IS NULL
			AND retry_count < ?
			AND (
				last_attempted_at IS NULL
				OR (retry_count = 1 AND last_attempted_at < now() - interval '30 seconds')
				OR (retry_count = 2 AND last_attempted_at < now() - interval '2 minutes')
				OR (retry_count = 3 AND last_attempted_at < now() - interval '10 minutes')
				OR (retry_count = 4 AND last_attempted_at < now() - interval '30 minutes')
				OR (retry_count >= 5 AND last_attempted_at < now() - interval '1 hour')
			)
		`, outboxMaxRetry).
		Order("last_attempted_at NULLS FIRST, id ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf(
			"outbox: pending batch: %w", err,
		)
	}
	return rows, nil
}

const outboxMaxRetry = 10

const (
	drainBreakerThreshold = 3
	drainBreakerCooldown  = 60 * time.Second
)
