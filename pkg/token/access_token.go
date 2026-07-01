// Package token implements user-facing Opaque Access Tokens (OAT)
// and refresh tokens for App authentication.
//
// Access tokens are random 32-byte values base64url-encoded and
// prefixed with "vg_at_". Their SHA-256 hash is the Redis key
// pointing at the JSON-encoded AccessTokenClaims; the raw token
// never appears in Redis. TTL is configured per environment
// (default 30 minutes) and validation is a single Redis GET.
//
// Refresh tokens follow the same shape with prefix "vg_rt_" and
// only their SHA-256 hex hash is persisted (in auth.refresh_tokens).
// TTL is configured per environment (default 90 days).
//
// This package replaces user-facing JWTs. The pkg/jwt package is
// reserved for admin Service-JWTs and MUST NOT be imported on
// user-facing paths.
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
	// accessTokenPrefix tags every raw access token so callers
	// can distinguish OAT strings from other secrets at a glance.
	accessTokenPrefix = "vg_at_"
	// refreshTokenPrefix tags every raw refresh token.
	refreshTokenPrefix = "vg_rt_"
	// tokenByteLen is the random byte count behind each token.
	// 32 bytes -> 256 bits of entropy, well above brute-force
	// reach for the configured TTL.
	tokenByteLen = 32
)

// ErrTokenNotFound is returned by Validate when the token is
// absent from Redis (expired, revoked, or never issued).
var ErrTokenNotFound = errors.New("access token not found or expired")

// AccessTokenClaims is the payload stored in Redis for each
// live access token. Returned by Validate so middleware can
// populate the request context without a DB round-trip.
type AccessTokenClaims struct {
	// SessionID points at the auth.device_sessions row.
	SessionID int64 `json:"session_id"`
	// UserID is the authenticated user (auth.users.id).
	UserID int64 `json:"user_id"`
	// DeviceID is the per-device session identifier; the auth
	// middleware compares it against the X-Device-Id header.
	DeviceID string `json:"device_id"`
	// IssuedAt is the UTC timestamp set at Generate time.
	IssuedAt time.Time `json:"issued_at"`
}

// AccessTokenManager manages opaque access token lifecycle
// via Redis. Used exclusively for user-facing authentication.
type AccessTokenManager interface {
	// Generate issues a new opaque access token, stores its
	// claims in Redis under SHA-256(token) with the configured
	// TTL, and returns the raw token string for the client.
	Generate(
		ctx context.Context,
		sessionID int64,
		userID int64,
		deviceID string,
	) (string, error)
	// Validate looks the token up in Redis. Returns
	// ErrTokenNotFound when the key is missing or expired.
	Validate(
		ctx context.Context,
		token string,
	) (*AccessTokenClaims, error)
	// Revoke deletes the Redis key. Safe to call on a token
	// that no longer exists.
	Revoke(ctx context.Context, token string) error
}

type redisAccessTokenManager struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisAccessTokenManager returns an AccessTokenManager
// backed by Redis. The ttl argument controls the lifetime of
// every token Generate issues.
func NewRedisAccessTokenManager(
	client *redis.Client,
	ttl time.Duration,
) AccessTokenManager {
	return &redisAccessTokenManager{client: client, ttl: ttl}
}

// redisKey derives the Redis key from a raw token using SHA-256.
// The raw token never appears as a Redis key — only its hash does.
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
	// Draw the token's entropy from the OS CSPRNG. A failure
	// here aborts authentication rather than emitting a weak
	// token.
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

// GenerateRefreshToken creates a new opaque refresh token and
// returns both the raw token (sent to the client) and its
// SHA-256 hex hash (persisted in auth.refresh_tokens.token_hash).
// The raw token is never stored server-side; revocation and
// rotation operate on the hash.
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
