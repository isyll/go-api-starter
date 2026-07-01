package cache

import (
	"context"
	"fmt"
	"strconv"
)

// Tag conventions: snake_case, lowercase, colon-separated. Each
// constructor names one kind of dependency shared by a cache writer
// and an invalidator.

// UserTag groups everything scoped to a single user (profile,
// settings). Drop it when the user changes.
func UserTag(userID int64) string {
	return "user:" + strconv.FormatInt(userID, 10)
}

// InvalidateByTags drops every cache key tracked by the given tags
// in a single Redis pipeline. Empty input is a no-op. It returns the
// first Redis error but still attempts every operation.
func (c *CacheManager) InvalidateByTags(
	ctx context.Context,
	tags ...string,
) error {
	if len(tags) == 0 {
		return nil
	}

	tagKeys := make([]string, 0, len(tags))
	for _, t := range tags {
		tagKeys = append(tagKeys, c.buildTagKey(t))
	}

	allKeys := make(map[string]struct{})
	for _, tagKey := range tagKeys {
		members, err := c.client.SMembers(ctx, tagKey).Result()
		if err != nil {
			return fmt.Errorf("cache: read tag %s: %w", tagKey, err)
		}
		for _, k := range members {
			allKeys[k] = struct{}{}
		}
	}

	pipe := c.client.Pipeline()
	for k := range allKeys {
		pipe.Del(ctx, k)
	}
	for _, tagKey := range tagKeys {
		pipe.Del(ctx, tagKey)
	}
	_, err := pipe.Exec(ctx)
	return err
}
