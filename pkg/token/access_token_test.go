package token

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (AccessTokenManager, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewRedisAccessTokenManager(client, 30*time.Minute), mr
}

func TestGenerateValidateRoundTrip(t *testing.T) {
	mgr, _ := setup(t)
	ctx := context.Background()

	tok, err := mgr.Generate(ctx, 7, 42, "dev-1")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(tok, "at_"))

	claims, err := mgr.Validate(ctx, tok)
	require.NoError(t, err)
	assert.Equal(t, int64(7), claims.SessionID)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "dev-1", claims.DeviceID)
}

func TestValidateUnknownToken(t *testing.T) {
	mgr, _ := setup(t)
	_, err := mgr.Validate(context.Background(), "at_bogus")
	assert.ErrorIs(t, err, ErrTokenNotFound)
}

func TestRevokeInvalidatesToken(t *testing.T) {
	mgr, _ := setup(t)
	ctx := context.Background()

	tok, err := mgr.Generate(ctx, 7, 42, "dev-1")
	require.NoError(t, err)
	require.NoError(t, mgr.Revoke(ctx, tok))

	_, err = mgr.Validate(ctx, tok)
	assert.ErrorIs(t, err, ErrTokenNotFound)
}

func TestTokenExpires(t *testing.T) {
	mgr, mr := setup(t)
	ctx := context.Background()

	tok, err := mgr.Generate(ctx, 7, 42, "dev-1")
	require.NoError(t, err)

	mr.FastForward(31 * time.Minute)
	_, err = mgr.Validate(ctx, tok)
	assert.ErrorIs(t, err, ErrTokenNotFound)
}

func TestTokensAreUniqueAndHashedAtRest(t *testing.T) {
	mgr, mr := setup(t)
	ctx := context.Background()

	a, err := mgr.Generate(ctx, 1, 1, "d")
	require.NoError(t, err)
	b, err := mgr.Generate(ctx, 1, 1, "d")
	require.NoError(t, err)
	assert.NotEqual(t, a, b)

	// The raw token must never appear as a Redis key: only its hash.
	for _, key := range mr.Keys() {
		assert.NotContains(t, key, a)
		assert.NotContains(t, key, b)
	}
}
