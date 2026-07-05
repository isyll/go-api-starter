// Package token implements opaque access and refresh tokens.
package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	accessTokenPrefix  = "at_"
	refreshTokenPrefix = "rt_"
	tokenByteLen       = 32
)

var ErrTokenNotFound = errors.New("access token not found or expired")

type AccessTokenClaims struct {
	SessionID int64     `json:"session_id"`
	UserID    int64     `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	IssuedAt  time.Time `json:"issued_at"`
}

type AccessTokenManager interface {
	Generate(
		ctx context.Context,
		sessionID int64,
		userID int64,
		deviceID string,
	) (string, error)
	Validate(
		ctx context.Context,
		token string,
	) (*AccessTokenClaims, error)
	Revoke(ctx context.Context, token string) error
}

type redisAccessTokenManager struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisAccessTokenManager(
	client *redis.Client,
	ttl time.Duration,
) AccessTokenManager {
	return &redisAccessTokenManager{client: client, ttl: ttl}
}

func redisKey(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return "at:" + hex.EncodeToString(h[:])
}

func (m *redisAccessTokenManager) Generate(
	ctx context.Context,
	sessionID int64,
	userID int64,
	deviceID string,
) (string, error) {
	b := make([]byte, tokenByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("token entropy failure: %w", err)
	}
	rawToken := accessTokenPrefix +
		base64.URLEncoding.WithPadding(base64.NoPadding).
			EncodeToString(b)

	claims := AccessTokenClaims{
		SessionID: sessionID,
		UserID:    userID,
		DeviceID:  deviceID,
		IssuedAt:  time.Now().UTC(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token claims: %w", err)
	}

	if err := m.client.Set(ctx,
		redisKey(rawToken), payload,
		m.ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to store access token: %w", err)
	}
	return rawToken, nil
}

func (m *redisAccessTokenManager) Validate(
	ctx context.Context,
	rawToken string,
) (*AccessTokenClaims, error) {
	data, err := m.client.Get(ctx, redisKey(rawToken)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve access token: %w",
			err,
		)
	}

	var claims AccessTokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return nil, fmt.Errorf("malformed token payload: %w", err)
	}
	return &claims, nil
}

func (m *redisAccessTokenManager) Revoke(
	ctx context.Context,
	rawToken string,
) error {
	return m.client.Del(ctx, redisKey(rawToken)).Err()
}

func GenerateRefreshToken() (rawToken, tokenHash string, err error) {
	b := make([]byte, tokenByteLen)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("token entropy failure: %w", err)
	}
	rawToken = refreshTokenPrefix +
		base64.URLEncoding.WithPadding(base64.NoPadding).
			EncodeToString(b)
	h := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(h[:])
	return rawToken, tokenHash, nil
}
