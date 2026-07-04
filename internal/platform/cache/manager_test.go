package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCacheManager(
	t *testing.T,
) (*CacheManager, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewCacheManager(client, "test"), mr
}

func TestCacheManager_SetAndGet(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	t.Run("stores and retrieves a value", func(t *testing.T) {
		type sample struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		err := mgr.Set(
			ctx, "user:1",
			sample{Name: "Alice", Age: 30}, CacheShort,
		)
		require.NoError(t, err)

		var got sample
		found, err := mgr.Get(ctx, "user:1", &got)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "Alice", got.Name)
		assert.Equal(t, 30, got.Age)
	})

	t.Run("returns false when key missing", func(t *testing.T) {
		var got map[string]any
		found, err := mgr.Get(ctx, "nonexistent", &got)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestCacheManager_SetWithTags(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	opts := CacheOptions{
		TTL:  15 * time.Minute,
		Tags: []string{"users", "profiles"},
	}
	err := mgr.Set(ctx, "user:tagged", "data", opts)
	require.NoError(t, err)

	var got string
	found, err := mgr.Get(ctx, "user:tagged", &got)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "data", got)
}

func TestCacheManager_Delete(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	_ = mgr.Set(ctx, "to-delete", "value", CacheShort)

	err := mgr.Delete(ctx, "to-delete")
	require.NoError(t, err)

	var got string
	found, err := mgr.Get(ctx, "to-delete", &got)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestCacheManager_GetOrSet(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	t.Run("fetches and caches on miss", func(t *testing.T) {
		var result string
		fromCache, err := mgr.GetOrSet(
			ctx, "lazy:1", &result,
			func() (any, error) { return "computed", nil },
			CacheShort,
		)
		require.NoError(t, err)
		assert.False(t, fromCache)
		assert.Equal(t, "computed", result)
	})

	t.Run("returns cached value on hit", func(t *testing.T) {
		var result string
		fromCache, err := mgr.GetOrSet(
			ctx, "lazy:1", &result,
			func() (any, error) { return "should-not-reach", nil },
			CacheShort,
		)
		require.NoError(t, err)
		assert.True(t, fromCache)
		assert.Equal(t, "computed", result)
	})
}

func TestCacheManager_InvalidateByTag(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	opts := CacheOptions{
		TTL:  15 * time.Minute,
		Tags: []string{"group-a"},
	}

	_ = mgr.Set(ctx, "a:1", "v1", opts)
	_ = mgr.Set(ctx, "a:2", "v2", opts)

	err := mgr.InvalidateByTag(ctx, "group-a")
	require.NoError(t, err)

	var got string
	found, _ := mgr.Get(ctx, "a:1", &got)
	assert.False(t, found)

	found, _ = mgr.Get(ctx, "a:2", &got)
	assert.False(t, found)
}

func TestCacheManager_InvalidatePattern(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	_ = mgr.Set(ctx, "items:1", "a", CacheShort)
	_ = mgr.Set(ctx, "items:2", "b", CacheShort)
	_ = mgr.Set(ctx, "other:1", "c", CacheShort)

	err := mgr.InvalidatePattern(ctx, "items:*")
	require.NoError(t, err)

	var got string
	found, _ := mgr.Get(ctx, "items:1", &got)
	assert.False(t, found)
	found, _ = mgr.Get(ctx, "items:2", &got)
	assert.False(t, found)

	found, _ = mgr.Get(ctx, "other:1", &got)
	assert.True(t, found)
}

func TestCacheManager_Flush(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	_ = mgr.Set(ctx, "f1", "v1", CacheShort)
	_ = mgr.Set(ctx, "f2", "v2", CacheShort)

	err := mgr.Flush(ctx)
	require.NoError(t, err)

	var got string
	found, _ := mgr.Get(ctx, "f1", &got)
	assert.False(t, found)
	found, _ = mgr.Get(ctx, "f2", &got)
	assert.False(t, found)
}

func TestCacheManager_StoreAndGetTokenData(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	data := map[string]any{
		"user_id": float64(42),
		"scope":   "read_write",
	}

	err := mgr.StoreTokenData(
		ctx, "token:abc", data, 10*time.Minute,
	)
	require.NoError(t, err)

	result, found, err := mgr.GetTokenData(ctx, "token:abc")
	require.NoError(t, err)
	assert.True(t, found)
	assert.InDelta(t, float64(42), result["user_id"], 0.01)
	assert.Equal(t, "read_write", result["scope"])
}

func TestCacheManager_GetTokenData_Missing(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	result, found, err := mgr.GetTokenData(ctx, "token:missing")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestCacheManager_DeleteTokenData(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	ctx := context.Background()

	_ = mgr.StoreTokenData(
		ctx, "token:del",
		map[string]any{"x": "y"}, 10*time.Minute,
	)

	err := mgr.DeleteTokenData(ctx, "token:del")
	require.NoError(t, err)

	_, found, err := mgr.GetTokenData(ctx, "token:del")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestCacheManager_CacheExpiry(t *testing.T) {
	mgr, mr := setupCacheManager(t)
	ctx := context.Background()

	opts := CacheOptions{TTL: 2 * time.Second}
	_ = mgr.Set(ctx, "expire-me", "value", opts)

	mr.FastForward(3 * time.Second)

	var got string
	found, err := mgr.Get(ctx, "expire-me", &got)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestCacheManager_BuildKey(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	key := mgr.buildKey("user:42")
	assert.Equal(t, "test:cache:user:42", key)
}

func TestCacheManager_BuildTagKey(t *testing.T) {
	mgr, _ := setupCacheManager(t)
	key := mgr.buildTagKey("users")
	assert.Equal(t, "test:tag:users", key)
}
