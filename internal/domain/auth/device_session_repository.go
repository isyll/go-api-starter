package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// Column names for the auth.device_sessions and
	// auth.refresh_tokens GORM Updates maps. Centralized so a
	// schema rename is a one-line change.
	colRevokedAt     = "revoked_at"
	colRevokedReason = "revoked_reason"
	colCreatedAt     = "created_at"
)

// DeviceSessionRepository defines the data-access contract for
// auth.device_sessions rows.
type DeviceSessionRepository interface {
	// Create persists a new device session, panicking on unexpected DB failure.
	Create(
		ctx context.Context,
		session *models.DeviceSession,
	)
	// FindByID retrieves a device session by its primary key, returning an
	// error when absent.
	FindByID(
		ctx context.Context,
		id int64,
	) (*models.DeviceSession, error)
	// FindByUserAndDeviceID returns the active session for the given user and
	// device pair, or nil when none exists.
	FindByUserAndDeviceID(
		ctx context.Context,
		userID int64,
		deviceID string,
	) *models.DeviceSession
	// Revoke stamps revoked_at and the provided reason onto the session,
	// returning the updated record.
	Revoke(
		ctx context.Context,
		reason string,
		id int64,
	) (*models.DeviceSession, error)
	// FindAndUpdateActivity atomically loads a session and bumps
	// last_activity_at, returning the updated record.
	FindAndUpdateActivity(
		ctx context.Context,
		id int64,
	) (*models.DeviceSession, error)
	// UpdateActivity bumps last_activity_at for the given session without
	// loading the full record.
	UpdateActivity(ctx context.Context, id int64) error
	// CountActiveDevicesByUser returns the number of non-revoked sessions
	// whose last_activity_at falls within the inactivity window.
	CountActiveDevicesByUser(
		ctx context.Context,
		userID int64,
		inactivityTimeout time.Duration,
	) int64
	// FindActiveDevicesByUser returns non-revoked sessions whose
	// last_activity_at falls within the inactivity window.
	FindActiveDevicesByUser(
		ctx context.Context,
		userID int64,
		inactivityTimeout time.Duration,
	) []models.DeviceSession
	// RevokeAllByUserID revokes every active session for the given user,
	// recording the provided reason on each row.
	RevokeAllByUserID(
		ctx context.Context,
		userID int64,
		reason string,
	) error
}

type deviceSessionRepository struct {
	db *gorm.DB
}

func NewDeviceSessionRepository(db *gorm.DB) DeviceSessionRepository {
	return &deviceSessionRepository{
		db: db,
	}
}

func (r *deviceSessionRepository) Create(
	ctx context.Context,
	session *models.DeviceSession,
) {
	session.LastActivity = time.Now().UTC()
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		panic(fmt.Errorf(
			"failed to create device session: %w", err,
		))
	}
}

func (r *deviceSessionRepository) FindByID(
	ctx context.Context,
	id int64,
) (*models.DeviceSession, error) {
	var session models.DeviceSession

	err := r.db.WithContext(ctx).
		Preload(
			"User.ActiveSuspension",
			"is_permanent = ? OR (suspended_until "+
				"IS NOT NULL AND suspended_until > NOW())",
			true,
		).
		Preload("User.UserSettings").
		Preload("User").
		First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrSessionNotFound
		}
		panic(fmt.Errorf(
			"failed to get session %d: %w", id, err,
		))
	}

	return &session, nil
}

func (r *deviceSessionRepository) FindByUserAndDeviceID(
	ctx context.Context,
	userID int64,
	deviceID string,
) *models.DeviceSession {
	var session models.DeviceSession
	err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND device_id = ? AND revoked_at IS NULL",
			userID,
			deviceID,
		).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		panic(fmt.Errorf("failed to get user device session: %w", err))
	}
	return &session
}

func (r *deviceSessionRepository) Revoke(
	ctx context.Context,
	reason string,
	id int64,
) (*models.DeviceSession, error) {
	var session models.DeviceSession

	err := r.db.WithContext(ctx).
		Model(&session).
		Where("id = ?", id).
		Clauses(clause.Returning{}).
		Updates(map[string]any{
			colRevokedAt:     time.Now().UTC(),
			colRevokedReason: reason,
		}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrSessionNotFound
		}
		panic(fmt.Errorf("failed to revoke session %d: %w", id, err))
	}

	if session.ID == 0 {
		return nil, apperrors.ErrSessionNotFound
	}

	return &session, nil
}

func (r *deviceSessionRepository) FindAndUpdateActivity(
	ctx context.Context,
	id int64,
) (*models.DeviceSession, error) {
	var session models.DeviceSession

	err := r.db.WithContext(ctx).
		Model(&models.DeviceSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("last_activity", time.Now().UTC()).Error
	if err != nil {
		panic(fmt.Errorf(
			"failed to update session %d activity: %w", id, err,
		))
	}

	err = r.db.WithContext(ctx).
		Preload(
			"User.ActiveSuspension",
			"is_permanent = ? OR (suspended_until IS NOT NULL "+
				"AND suspended_until > NOW())",
			true,
		).
		Preload("User").
		First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrSessionNotFound
		}
		panic(fmt.Errorf(
			"failed to load session %d associations: %w", id, err,
		))
	}

	return &session, nil
}

func (r *deviceSessionRepository) UpdateActivity(
	ctx context.Context,
	id int64,
) error {
	return r.db.WithContext(ctx).
		Model(&models.DeviceSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("last_activity", time.Now().UTC()).
		Error
}

func (r *deviceSessionRepository) CountActiveDevicesByUser(
	ctx context.Context,
	userID int64,
	inactivityTimeout time.Duration,
) int64 {
	var count int64

	inactivityThreshold := time.Now().UTC().Add(-inactivityTimeout)

	err := r.db.WithContext(ctx).
		Model(&models.DeviceSession{}).
		Where(
			"user_id = ? AND revoked_at IS NULL AND last_activity > ?",
			userID, inactivityThreshold,
		).
		Count(&count).Error
	if err != nil {
		panic(fmt.Errorf("failed to count active devices: %w", err))
	}

	return count
}

func (r *deviceSessionRepository) FindActiveDevicesByUser(
	ctx context.Context,
	userID int64,
	inactivityTimeout time.Duration,
) []models.DeviceSession {
	var sessions []models.DeviceSession

	inactivityThreshold := time.Now().UTC().Add(-inactivityTimeout)

	err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND revoked_at IS NULL AND last_activity > ?",
			userID,
			inactivityThreshold,
		).
		Order("last_activity DESC").
		Find(&sessions).Error
	if err != nil {
		panic(fmt.Errorf("failed to find active devices: %w", err))
	}

	return sessions
}

func (r *deviceSessionRepository) RevokeAllByUserID(
	ctx context.Context,
	userID int64,
	reason string,
) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&models.DeviceSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{
			colRevokedAt:     now,
			colRevokedReason: reason,
		})

	if result.Error != nil {
		panic(fmt.Errorf(
			"revoke all sessions for user %d: %w",
			userID,
			result.Error,
		))
	}
	return nil
}
