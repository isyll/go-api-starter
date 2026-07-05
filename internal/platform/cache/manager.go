// Package cache wraps Redis for caching, tags, and token storage.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/isyll/go-grpc-starter/internal/metrics"
)

type CacheManager struct {
	client *redis.Client
	prefix string
	sf     singleflight.Group
}

type CacheOptions struct {
	TTL  time.Duration
	Tags []string
}

var (
	CacheLong = CacheOptions{TTL: 24 * time.Hour}

	CacheMedium = CacheOptions{TTL: 1 * time.Hour}

	CacheShort = CacheOptions{TTL: 15 * time.Minute}

	CacheTemp = CacheOptions{TTL: 2 * time.Minute}
)

func NewCacheManager(
	client *redis.Client,
	prefix string,
) *CacheManager {
	return &CacheManager{client: client, prefix: prefix}
}

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
		// Re-check inside the flight: a concurrent winner may have already
		// filled the cache between our miss and acquiring the flight.
		var raw json.RawMessage
		ok, gErr := c.Get(ctx, key, &raw)
		if gErr == nil && ok {
			return raw, nil
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

func singleflightKind(key string) string {
	if len(key) >= 8 && key[:8] == "session:" {
		return "session"
	}
	return "service"
}

func (c *CacheManager) Delete(ctx context.Context, key string) error {
	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// GetDel atomically reads and deletes a key. Use it for single-use tokens
// (email verification, password reset) so two concurrent requests cannot both
// consume the same token.
func (c *CacheManager) GetDel(
	ctx context.Context,
	key string,
	dest any,
) (bool, error) {
	data, err := c.client.GetDel(ctx, c.buildKey(key)).Bytes()
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

// InvalidatePattern deletes every key matching the glob pattern. It iterates
// with SCAN so it never blocks Redis the way KEYS would on a large keyspace.
func (c *CacheManager) InvalidatePattern(
	ctx context.Context,
	pattern string,
) error {
	return c.scanDelete(ctx, c.buildKey(pattern))
}

// Flush removes every cache entry under this manager's prefix.
func (c *CacheManager) Flush(ctx context.Context) error {
	return c.scanDelete(ctx, c.buildKey("*"))
}

func (c *CacheManager) scanDelete(ctx context.Context, pattern string) error {
	const scanCount = 200
	iter := c.client.Scan(ctx, 0, pattern, scanCount).Iterator()
	batch := make([]string, 0, scanCount)
	for iter.Next(ctx) {
		batch = append(batch, iter.Val())
		if len(batch) == scanCount {
			if err := c.client.Del(ctx, batch...).Err(); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(batch) > 0 {
		return c.client.Del(ctx, batch...).Err()
	}
	return nil
}

func (c *CacheManager) buildKey(key string) string {
	return fmt.Sprintf("%s:cache:%s", c.prefix, key)
}

func (c *CacheManager) buildTagKey(tag string) string {
	return fmt.Sprintf("%s:tag:%s", c.prefix, tag)
}
