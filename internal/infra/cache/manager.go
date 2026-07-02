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

	"github.com/isyll/go-api-starter/internal/metrics"
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

func newSameType(dest any) any {
	return dest
}

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

func (c *CacheManager) Delete(ctx context.Context, key string) error {
	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
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

func (c *CacheManager) InvalidateHTTPRoute(
	ctx context.Context,
	path string,
) error {
	pattern := fmt.Sprintf("http:*%s*", path)
	return c.InvalidatePattern(ctx, pattern)
}

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

func (c *CacheManager) DeleteTokenData(
	ctx context.Context,
	key string,
) error {
	return c.client.Del(ctx, key).Err()
}

func (c *CacheManager) UpdateTokenDataTTL(
	ctx context.Context,
	key string,
	ttl time.Duration,
) error {
	return c.client.Expire(ctx, key, ttl).Err()
}
