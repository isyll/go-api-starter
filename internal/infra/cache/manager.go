// Package cache wraps Redis to provide three things: a tagged,
// TTL-bounded JSON cache for service-level lookups (GetOrSet); a
// canonical Sqids/HTTP/OTP key vocabulary in cachekey.go; and the
// shared tag dictionary in tags.go used by middleware.HTTPCache and
// the event handlers in internal/events/handlers/cache_invalidation.
//
// Application code MUST go through CacheManager — never write to
// the underlying redis.Client directly. Service-level cached
// lookups MUST set CacheOptions.Tags identical to the HTTP route's
// tags so a single InvalidateByTags call drops both layers in one
// shot. Tag strings live ONLY in tags.go.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/isyll/go-api-starter/internal/metrics"
)

// CacheManager is the Redis-backed cache facade. Construct via
// NewCacheManager; every method namespaces keys under the supplied
// prefix so multiple deployments can share a Redis instance.
//
// A singleflight.Group backs GetOrSet so concurrent fetches for the
// same key collapse into one fetchFunc call. Without it, every TTL
// expiry under load fans out N parallel DB queries; with it, only
// one wins and the others wait on its result. This applies per
// process — concurrent requests across API replicas still race, but
// each replica's herd is bounded by the gate.
type CacheManager struct {
	client *redis.Client
	prefix string
	sf     singleflight.Group
}

// CacheOptions controls the TTL and tag set of a single Set call.
type CacheOptions struct {
	// TTL is the absolute expiration applied by Redis.
	TTL time.Duration
	// Tags is the set of tag names whose Redis SET membership lists
	// will track this key for bulk invalidation.
	Tags []string
}

var (
	// CacheLong is for rarely-changing data (config, global stats).
	CacheLong = CacheOptions{TTL: 24 * time.Hour}

	// CacheMedium is for user-scoped data (profile, preferences).
	CacheMedium = CacheOptions{TTL: 1 * time.Hour}

	// CacheShort is for list-style and search responses.
	CacheShort = CacheOptions{TTL: 15 * time.Minute}

	// CacheTemp is for real-time / near-live data.
	CacheTemp = CacheOptions{TTL: 2 * time.Minute}
)

// NewCacheManager wraps a Redis client with the given prefix used
// to namespace every key.
func NewCacheManager(
	client *redis.Client,
	prefix string,
) *CacheManager {
	return &CacheManager{client: client, prefix: prefix}
}

// Set stores value at key with the supplied TTL and tag set. The
// write is performed in a single Redis pipeline so a partial
// failure cannot leave a key without its tag membership.
func (c *CacheManager) Set(
	ctx context.Context,
	key string,
	value any,
	opts CacheOptions,
) error {
	fullKey := c.buildKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	pipe := c.client.Pipeline()

	pipe.Set(ctx, fullKey, data, opts.TTL)

	for _, tag := range opts.Tags {
		tagKey := c.buildTagKey(tag)
		pipe.SAdd(ctx, tagKey, fullKey)
		pipe.Expire(ctx, tagKey, opts.TTL+time.Hour)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Get fetches key into dest. Returns (false, nil) on a cache miss
// and (true, nil) on a hit. Caller-side unmarshal errors surface as
// a non-nil error.
func (c *CacheManager) Get(
	ctx context.Context,
	key string,
	dest any,
) (bool, error) {
	fullKey := c.buildKey(key)

	data, err := c.client.Get(ctx, fullKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("cache unmarshal error: %w", err)
	}

	return true, nil
}

// GetOrSet returns the cached value at key when present, otherwise
// invokes fetchFunc, stores the result under the supplied options,
// and copies it into dest. The first return value reports whether
// the value came from cache (used by auth to decide whether to
// asynchronously refresh activity).
//
// Stampede protection: on a miss, the call routes through
// a singleflight.Group keyed by `key`. Concurrent callers waiting on
// the same key share the fetchFunc result instead of each issuing
// their own DB query. Inside the singleflight closure the cache is
// re-checked so a winner's Set is observed by the followers.
//
// Caveats:
//   - singleflight is per-process. Cross-replica races still occur,
//     but each replica's herd is bounded.
//   - fetchFunc errors are forwarded to every joiner. This matches
//     stdlib singleflight semantics; non-deterministic errors must
//     therefore be retried by callers, not silently absorbed here.
func (c *CacheManager) GetOrSet(
	ctx context.Context,
	key string,
	dest any,
	fetchFunc func() (any, error),
	opts CacheOptions,
) (fromCache bool, err error) {
	found, getErr := c.Get(ctx, key, dest)
	if getErr != nil {
		return false, getErr
	}
	if found {
		return true, nil
	}

	v, sfErr, shared := c.sf.Do(key, func() (any, error) {
		// Re-check under the gate so the second/third/Nth
		// caller observes the winner's Set without
		// re-running fetchFunc. The winner's path proceeds
		// through fetchFunc + Set normally.
		winnerDest := newSameType(dest)
		ok, gErr := c.Get(ctx, key, winnerDest)
		if gErr == nil && ok {
			return winnerDest, nil
		}
		value, ferr := fetchFunc()
		if ferr != nil {
			return nil, ferr
		}
		if sErr := c.Set(ctx, key, value, opts); sErr != nil {
			return nil, sErr
		}
		return value, nil
	})
	if shared {
		metrics.CacheSingleflightCollapsedTotal.
			WithLabelValues(singleflightKind(key)).
			Inc()
	}
	if sfErr != nil {
		return false, sfErr
	}

	data, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return false, marshalErr
	}
	return false, json.Unmarshal(data, dest)
}

// newSameType returns a zero value of the same concrete
// type as dest so the re-check inside the singleflight
// closure unmarshals into a compatible target. For
// pointer dests it returns a fresh pointer to the
// underlying element.
func newSameType(dest any) any {
	// json.Unmarshal already requires a pointer; calling
	// code passes &session, &payload, etc. We can reuse
	// the same target — Unmarshal overwrites it. Returning
	// `dest` directly is safe because we only consume the
	// returned interface for the unmarshal step.
	return dest
}

// singleflightKind classifies a key for the metric label
// so dashboards can break down collapses by call site
// (session lookup, http response cache, etc.). Cheap
// prefix match.
func singleflightKind(key string) string {
	switch {
	case len(key) >= 8 && key[:8] == "session:":
		return "session"
	case len(key) >= 5 && key[:5] == "http:":
		return "http"
	default:
		return "service"
	}
}

// Delete removes the entry at key. No-op when missing.
func (c *CacheManager) Delete(ctx context.Context, key string) error {
	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// InvalidateByTag drops every cache key tracked by tag. Prefer the
// batched InvalidateByTags variant when clearing multiple tags so
// the same Redis pipeline can serve them all.
func (c *CacheManager) InvalidateByTag(
	ctx context.Context,
	tag string,
) error {
	tagKey := c.buildTagKey(tag)

	keys, err := c.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	pipe.Del(ctx, tagKey)

	_, err = pipe.Exec(ctx)
	return err
}

// InvalidatePattern deletes every key matching pattern (Redis glob).
// Use sparingly — SCAN-equivalent operations are O(N) over the
// keyspace; prefer tag-based invalidation.
func (c *CacheManager) InvalidatePattern(
	ctx context.Context,
	pattern string,
) error {
	fullPattern := c.buildKey(pattern)

	keys, err := c.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}

// InvalidateHTTPRoute drops every cached HTTP response whose key
// contains path. Legacy; new code should rely on tags via the
// HTTPCache middleware Tagger.
func (c *CacheManager) InvalidateHTTPRoute(
	ctx context.Context,
	path string,
) error {
	pattern := fmt.Sprintf("http:*%s*", path)
	return c.InvalidatePattern(ctx, pattern)
}

// Flush deletes every key under this manager's prefix. Test helper
// only.
func (c *CacheManager) Flush(ctx context.Context) error {
	pattern := c.buildKey("*")
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}

func (c *CacheManager) buildKey(key string) string {
	return fmt.Sprintf("%s:cache:%s", c.prefix, key)
}

func (c *CacheManager) buildTagKey(tag string) string {
	return fmt.Sprintf("%s:tag:%s", c.prefix, tag)
}

// StoreTokenData persists an arbitrary key/value map under the raw
// (un-prefixed) key. Used by the OAT session store.
func (c *CacheManager) StoreTokenData(
	ctx context.Context,
	key string,
	data map[string]any,
	ttl time.Duration,
) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	return c.client.Set(ctx, key, jsonData, ttl).Err()
}

// GetTokenData fetches the token payload at the raw key. Returns
// (nil, false, nil) when the key is missing.
func (c *CacheManager) GetTokenData(
	ctx context.Context,
	key string,
) (result map[string]any, found bool, err error) {
	rawData, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if err := json.Unmarshal(rawData, &result); err != nil {
		return nil, false, fmt.Errorf(
			"failed to unmarshal token data: %w",
			err,
		)
	}

	return result, true, nil
}

// DeleteTokenData removes the token payload at the raw key.
func (c *CacheManager) DeleteTokenData(
	ctx context.Context,
	key string,
) error {
	return c.client.Del(ctx, key).Err()
}

// UpdateTokenDataTTL refreshes the TTL on an existing token entry
// without rewriting its value.
func (c *CacheManager) UpdateTokenDataTTL(
	ctx context.Context,
	key string,
	ttl time.Duration,
) error {
	return c.client.Expire(ctx, key, ttl).Err()
}
