package event

import (
	"time"

	"github.com/hibiken/asynq"
)

type SubscribeOption func(*subConfig)

type subConfig struct {
	critical  bool
	asyncOpts []asynq.Option
	taskIDFn  func(Event) string
}

func newSubConfig(opts []SubscribeOption) subConfig {
	cfg := subConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func WithCritical() SubscribeOption {
	return func(c *subConfig) { c.critical = true }
}

func WithQueue(name string) SubscribeOption {
	return func(c *subConfig) {
		c.asyncOpts = append(c.asyncOpts, asynq.Queue(name))
	}
}

func WithUniqueWindow(d time.Duration) SubscribeOption {
	return func(c *subConfig) {
		c.asyncOpts = append(c.asyncOpts, asynq.Unique(d))
	}
}

func WithTaskIDFn(fn func(Event) string) SubscribeOption {
	return func(c *subConfig) { c.taskIDFn = fn }
}

func WithMaxRetry(n int) SubscribeOption {
	return func(c *subConfig) {
		c.asyncOpts = append(c.asyncOpts, asynq.MaxRetry(n))
	}
}

func WithTimeout(d time.Duration) SubscribeOption {
	return func(c *subConfig) {
		c.asyncOpts = append(c.asyncOpts, asynq.Timeout(d))
	}
}
