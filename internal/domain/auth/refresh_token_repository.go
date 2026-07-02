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

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken)
	FindByTokenHash(
		ctx context.Context,
		tokenHash string,
	) (*models.RefreshToken, error)
	RevokeByTokenHash(
		ctx context.Context,
		tokenHash, reason string,
	) error
	RevokeBySessionID(
		ctx context.Context,
		sessionID int64,
		reason string,
	) error
	RevokeByTokenFamily(
		ctx context.Context,
		tokenFamily, reason string,
	) error
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
