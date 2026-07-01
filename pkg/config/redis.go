package config

import (
	"fmt"
	"time"
)

// RedisConfig holds the connection pool parameters, cache key
// prefixes, and session key settings loaded from configs/redis.yaml.
type RedisConfig struct {
	// Connection holds the raw dial parameters for the Redis client.
	Connection struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		// DB is the Redis logical database index (0–15).
		DB           int `yaml:"db"`
		PoolSize     int `yaml:"pool_size"`
		MinIdleConns int `yaml:"min_idle_conns"`
		MaxIdleConns int `yaml:"max_idle_conns"`
		// ConnMaxLifetime is the maximum time a connection may be reused
		// before it is replaced.
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
		// ConnMaxIdleTime is how long an idle connection is kept before
		// being closed.
		ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
		// PoolTimeout is how long the client waits for a free connection
		// before returning an error.
		PoolTimeout time.Duration `yaml:"pool_timeout"`
		// MaxRetries is the number of command retries on network errors.
		MaxRetries      int           `yaml:"max_retries"`
		MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`
		MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`
		DialTimeout     time.Duration `yaml:"dial_timeout"`
		ReadTimeout     time.Duration `yaml:"read_timeout"`
		WriteTimeout    time.Duration `yaml:"write_timeout"`
	} `yaml:"redis"`
	// Cache holds the key-prefix strings used to namespace cached
	// payloads and avoid cross-purpose collisions.
	Cache struct {
		Prefix      string `yaml:"prefix"`
		UserSession string `yaml:"user_session"`
		UserProfile string `yaml:"user_profile"`
		TripSearch  string `yaml:"trip_search"`
		RateLimit   string `yaml:"rate_limit"`
		OTP         string `yaml:"otp"`
	} `yaml:"cache"`
	// Session holds the key prefix and expiry for opaque access
	// token entries written to Redis.
	Session struct {
		KeyPrefix string `yaml:"key_prefix"`
		// Expiry is the Redis TTL for each session key expressed as a
		// duration string (e.g. "30m").
		Expiry string `yaml:"expiry"`
	} `yaml:"session"`
}

// Credentials returns the dial address and password for the Redis
// connection, suitable for passing directly to redis.NewClient.
func (r *RedisConfig) Credentials() (addr, password string) {
	addr = fmt.Sprintf("%s:%d", r.Connection.Host, r.Connection.Port)
	password = r.Connection.Password
	return
}
