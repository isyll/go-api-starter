package config

import (
	"fmt"
	"time"
)

type RedisConfig struct {
	Connection struct {
		Host            string        `yaml:"host"`
		Port            int           `yaml:"port"`
		Password        string        `yaml:"password"`
		DB              int           `yaml:"db"`
		PoolSize        int           `yaml:"pool_size"`
		MinIdleConns    int           `yaml:"min_idle_conns"`
		MaxIdleConns    int           `yaml:"max_idle_conns"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
		ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
		PoolTimeout     time.Duration `yaml:"pool_timeout"`
		MaxRetries      int           `yaml:"max_retries"`
		MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`
		MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`
		DialTimeout     time.Duration `yaml:"dial_timeout"`
		ReadTimeout     time.Duration `yaml:"read_timeout"`
		WriteTimeout    time.Duration `yaml:"write_timeout"`
	} `yaml:"redis"`
	Cache struct {
		Prefix      string `yaml:"prefix"`
		UserSession string `yaml:"user_session"`
		UserProfile string `yaml:"user_profile"`
		TripSearch  string `yaml:"trip_search"`
		RateLimit   string `yaml:"rate_limit"`
		OTP         string `yaml:"otp"`
	} `yaml:"cache"`
	Session struct {
		KeyPrefix string `yaml:"key_prefix"`
		Expiry    string `yaml:"expiry"`
	} `yaml:"session"`
}

func (r *RedisConfig) Credentials() (addr, password string) {
	addr = fmt.Sprintf("%s:%d", r.Connection.Host, r.Connection.Port)
	password = r.Connection.Password
	return
}
