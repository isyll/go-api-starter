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

// ErrOutboxDuplicate is returned by Write/Publish when a Dedupable
// event collides with an existing pending row on the partial unique
// index ux_outbox_pending_dedupe. The caller can treat this as a
// successful no-op: the original row will be processed by the drain.
var ErrOutboxDuplicate = errors.New(
	"outbox: duplicate pending row with same (event_type, dedupe_key)",
)

// OutboxEvent is the DB row for the transactional outbox. Unprocessed
// rows are retried by the drain goroutine so that a process crash between
// domain commit and Asynq enqueue does not silently drop async event
// handlers.
//
// The bus writes the outbox row inside the originating mutation's
// transaction via Bus.Publish + persistence.Manager.WithTx.
// This eliminates the window between domain commit and outbox insert
// — the row and the mutation share commit fate.
type OutboxEvent struct {
	// ID is the int64 primary key of the outbox row.
	ID int64 `gorm:"primaryKey"`
	// EventType is the bus routing key (e.g. "user.account_deleted"); maps
	// back to a Factory via registry.FactoryFor.
	EventType string `gorm:"not null"`
	// Payload is the JSON-encoded Event value.
	Payload json.RawMessage `gorm:"type:jsonb;not null"`
	// DedupeKey is an optional caller-supplied idempotency key.
	// When non-nil, a partial unique index (ux_outbox_pending_dedupe)
	// blocks a second pending row with the same (EventType, DedupeKey)
	// from being inserted. nil opts out: every publish writes a
	// fresh row, which is the right default for non-idempotent
	// producers.
	DedupeKey *string
	// CreatedAt is the insertion timestamp.
	CreatedAt time.Time `gorm:"not null;default:now()"`
	// ProcessedAt is set when the dispatcher confirms enqueue (or the
	// row is dead-lettered).
	ProcessedAt *time.Time
	// RetryCount is the number of unsuccessful drain attempts.
	RetryCount int `gorm:"not null;default:0"`
	// LastError is the most recent failure message (truncated).
	LastError *string
	// LastAttemptedAt is the timestamp of the most recent drain
	// attempt. Combined with RetryCount it gates the next attempt
	// via outboxBackoffFor. NULL means "never attempted" — drain
	// ASAP.
	LastAttemptedAt *time.Time
}

// TableName returns the schema-qualified table name "events.outbox".
func (OutboxEvent) TableName() string { return "events.outbox" }

// OutboxDeadLetterEvent holds outbox rows that exceeded the retry limit
// or could not be processed due to an unknown event type or unmarshal
// failure.
type OutboxDeadLetterEvent struct {
	// ID is the dead-letter row primary key.
	ID int64 `gorm:"primaryKey"`
	// SourceID is the original events.outbox row ID.
	SourceID int64 `gorm:"not null"`
	// EventType is the bus routing key, preserved for triage.
	EventType string `gorm:"not null"`
	// Payload is the original JSON event payload.
	Payload json.RawMessage `gorm:"type:jsonb;not null"`
	// FailureReason is the machine-readable category (e.g.
	// "retry_exhausted", "unknown_event_type", "unmarshal_failed").
	FailureReason string `gorm:"not null"`
	// LastError is the human-readable error text, if available.
	LastError *string
	// CreatedAt is the dead-letter insertion timestamp.
	CreatedAt time.Time `gorm:"not null;default:now()"`
	// FailedAt is the moment the drain decided to dead-letter.
	FailedAt time.Time `gorm:"not null;default:now()"`
}

// TableName returns the schema-qualified table name
// "events.outbox_dead_letter".
func (OutboxDeadLetterEvent) TableName() string {
	return "events.outbox_dead_letter"
}

// OutboxRepository handles persistence for outbox rows.
type OutboxRepository struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewOutboxRepository creates an OutboxRepository.
func NewOutboxRepository(
	db *gorm.DB,
	logx *logger.Logger,
) *OutboxRepository {
	return &OutboxRepository{db: db, logger: logx}
}

// Write inserts a pending outbox row. When called inside a
// persistence.Manager.WithTx closure, the insert joins the same
// database transaction as the calling mutation.
//
// When evt implements Dedupable and returns a non-empty key, a
// duplicate pending row triggers ErrOutboxDuplicate instead of a
// generic insert failure — the caller can treat the duplicate as
// a successful no-op since the original row is still being
// processed.
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

// isOutboxDedupeViolation reports whether err is a PostgreSQL
// unique_violation (SQLSTATE 23505) on the ux_outbox_pending_dedupe
// index. Any other unique violation falls through to the generic
// error path so it stays visible.
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

// Publish writes an outbox row and logs failures. It is the preferred
// call site inside persistence.Manager.WithTx closures: the row is
// committed atomically with the mutation, and any write failure rolls
// back the entire transaction.
//
// A Dedupable duplicate (ErrOutboxDuplicate) is treated as a
// successful no-op so callers do not have to special-case it: the
// original row already exists in the queue and will be processed by
// the drain.
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

// MarkProcessed stamps the row as successfully dispatched.
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

// MarkFailed increments the retry counter, stamps last_attempted_at
// to NOW(), and records the last error. The next drain pass uses
// last_attempted_at + outboxBackoffFor(retry_count) to decide
// whether the row is ready to retry — under a sustained Redis
// outage this self-throttles the drain instead of hammering
// Postgres on every tick.
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

// DeadLetter moves a row to events.outbox_dead_letter and marks the
// source row as processed so the drainer does not retry it.
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

// PendingBatch returns up to limit unprocessed rows whose backoff
// window has elapsed, oldest first. Uses FOR UPDATE SKIP LOCKED so
// multiple drainer instances can run concurrently without
// delivering the same row twice.
//
// Backoff predicate (in SQL):
//
//	last_attempted_at IS NULL                              -- never tried
//	OR retry_count = 1 AND last_attempted_at < now() - 30s
//	OR retry_count = 2 AND last_attempted_at < now() - 2m
//	...
//
// This mirrors outboxBackoffFor exactly. Adjust both in lockstep.
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

// Drain-level circuit breaker. When drainOnce sees this many
// consecutive total-failure passes (every fetched row failed to
// publish, typically because Asynq/Redis is unreachable), the
// breaker opens for drainBreakerCooldown — the next drain tick is
// observed but its body is skipped, so Redis is not hammered with
// per-row publish attempts during the outage. A successful pass
// closes the breaker and resets the counter.
const (
	drainBreakerThreshold = 3
	drainBreakerCooldown  = 60 * time.Second
)
