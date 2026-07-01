package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	models "github.com/isyll/go-api-starter/internal/models"

	"gorm.io/gorm"
)

// RefreshTokenRepository defines the data-access contract for
// auth.refresh_tokens rows.
type RefreshTokenRepository interface {
	// Create persists a new refresh-token record, panicking on unexpected
	// DB failure.
	Create(ctx context.Context, token *models.RefreshToken)
	// FindByTokenHash looks up a refresh token by its SHA-256 hash, returning
	// ErrInvalidToken when absent.
	FindByTokenHash(
		ctx context.Context,
		tokenHash string,
	) (*models.RefreshToken, error)
	// RevokeByTokenHash marks the token identified by its hash as revoked
	// with the given reason.
	RevokeByTokenHash(
		ctx context.Context,
		tokenHash, reason string,
	) error
	// RevokeBySessionID revokes all refresh tokens belonging to the given
	// device session.
	RevokeBySessionID(
		ctx context.Context,
		sessionID int64,
		reason string,
	) error
	// RevokeByTokenFamily revokes all tokens that share the given family
	// identifier, invalidating a replay-attack lineage.
	RevokeByTokenFamily(
		ctx context.Context,
		tokenFamily, reason string,
	) error
	// CleanupExpired deletes refresh-token rows past their expiry date and
	// returns the count of deleted records.
	CleanupExpired(ctx context.Context) (int64, error)
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(
	ctx context.Context,
	token *models.RefreshToken,
) {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		panic(fmt.Errorf(
			"failed to create refresh token record with value %v: %w",
			token,
			err,
		))
	}
}

func (r *refreshTokenRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*models.RefreshToken, error) {
	if len(tokenHash) < 8 {
		return nil, ErrInvalidToken
	}
	var token models.RefreshToken
	// Narrow via the indexed prefix first, then validate the full hash
	// in constant time to prevent timing-side-channel token enumeration.
	err := r.db.WithContext(ctx).
		Preload("Session.User").
		Where("token_prefix = ?", tokenHash[:8]).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		panic(fmt.Errorf("find refresh token: %w", err))
	}
	if subtle.ConstantTimeCompare(
		[]byte(token.TokenHash),
		[]byte(tokenHash),
	) != 1 {
		return nil, ErrInvalidToken
	}
	return &token, nil
}

func (r *refreshTokenRepository) RevokeByTokenHash(
	ctx context.Context,
	tokenHash string,
	reason string,
) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Updates(map[string]any{
			colRevokedAt:     now,
			colRevokedReason: reason,
		})

	if result.Error != nil {
		panic(fmt.Errorf("revoke token: %w", result.Error))
	}
	return nil
}

func (r *refreshTokenRepository) RevokeBySessionID(
	ctx context.Context,
	sessionID int64,
	reason string,
) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{
			colRevokedAt:     now,
			colRevokedReason: reason,
		})

	if result.Error != nil {
		panic(fmt.Errorf(
			"revoke tokens by session: %w", result.Error,
		))
	}
	return nil
}

func (r *refreshTokenRepository) RevokeByTokenFamily(
	ctx context.Context,
	tokenFamily string,
	reason string,
) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where(
			"token_family = ? AND revoked_at IS NULL",
			tokenFamily,
		).
		Updates(map[string]any{
			colRevokedAt:     now,
			colRevokedReason: reason,
		})

	if result.Error != nil {
		panic(fmt.Errorf(
			"revoke tokens by family: %w", result.Error,
		))
	}
	return nil
}

func (r *refreshTokenRepository) CleanupExpired(
	ctx context.Context,
) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ? AND revoked_at IS NULL", time.Now().UTC()).
		Delete(&models.RefreshToken{})

	if result.Error != nil {
		panic(fmt.Errorf(
			"cleanup expired tokens: %w", result.Error,
		))
	}
	return result.RowsAffected, nil
}
