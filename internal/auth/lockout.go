package auth

import (
	"context"
	"time"

	"github.com/isyll/go-grpc-starter/internal/platform/cache"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

const (
	defaultLockoutAttempts = 10
	defaultLockoutWindow   = 15 * time.Minute
)

// loginLimiter throttles password guessing per account with a Redis counter.
// It fails open: if Redis is unreachable the login proceeds, because locking
// every user out during a cache outage is the worse failure mode.
type loginLimiter struct {
	cache       *cache.CacheManager
	logger      *logger.Logger
	maxAttempts int
	window      time.Duration
}

func newLoginLimiter(
	cm *cache.CacheManager,
	logx *logger.Logger,
	maxAttempts int,
	window time.Duration,
) loginLimiter {
	if maxAttempts <= 0 {
		maxAttempts = defaultLockoutAttempts
	}
	if window <= 0 {
		window = defaultLockoutWindow
	}
	return loginLimiter{cache: cm, logger: logx, maxAttempts: maxAttempts, window: window}
}

func (l loginLimiter) blocked(ctx context.Context, email string) bool {
	n, err := l.cache.Counter(ctx, cache.RateLimitKey("login", email))
	if err != nil {
		l.logger.Warn("login lockout check failed; failing open", "error", err)
		return false
	}
	return n >= int64(l.maxAttempts)
}

func (l loginLimiter) recordFailure(ctx context.Context, email string) {
	if _, err := l.cache.IncrementWithTTL(
		ctx, cache.RateLimitKey("login", email), l.window,
	); err != nil {
		l.logger.Warn("login lockout increment failed", "error", err)
	}
}

func (l loginLimiter) reset(ctx context.Context, email string) {
	if err := l.cache.Delete(ctx, cache.RateLimitKey("login", email)); err != nil {
		l.logger.Warn("login lockout reset failed", "error", err)
	}
}
