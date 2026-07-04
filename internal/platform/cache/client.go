package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/isyll/go-grpc-starter/pkg/config"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf(
		"%s:%d",
		cfg.Connection.Host,
		cfg.Connection.Port,
	)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Connection.Password,
		DB:       cfg.Connection.DB,

		PoolSize:        cfg.Connection.PoolSize,
		MinIdleConns:    cfg.Connection.MinIdleConns,
		MaxIdleConns:    cfg.Connection.MaxIdleConns,
		ConnMaxLifetime: cfg.Connection.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Connection.ConnMaxIdleTime,
		PoolTimeout:     cfg.Connection.PoolTimeout,

		DialTimeout:  cfg.Connection.DialTimeout,
		ReadTimeout:  cfg.Connection.ReadTimeout,
		WriteTimeout: cfg.Connection.WriteTimeout,

		MaxRetries:      cfg.Connection.MaxRetries,
		MinRetryBackoff: cfg.Connection.MinRetryBackoff,
		MaxRetryBackoff: cfg.Connection.MaxRetryBackoff,
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
