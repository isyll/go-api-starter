package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// RegisterPoolCollectors registers pool gauges; call once after pools are ready.
func RegisterPoolCollectors(pool *pgxpool.Pool, rdb *redis.Client) {
	prometheus.MustRegister(
		&pgxPoolCollector{pool: pool},
		&redisPoolCollector{client: rdb},
	)
}

type pgxPoolCollector struct {
	pool *pgxpool.Pool
}

var (
	pgxTotalDesc = prometheus.NewDesc(
		"app_db_pool_total_conns", "Total connections in the pgx pool.", nil, nil)
	pgxIdleDesc = prometheus.NewDesc(
		"app_db_pool_idle_conns", "Idle connections in the pgx pool.", nil, nil)
	pgxAcquiredDesc = prometheus.NewDesc(
		"app_db_pool_acquired_conns", "Connections currently acquired.", nil, nil)
	pgxMaxDesc = prometheus.NewDesc(
		"app_db_pool_max_conns", "Configured maximum pool size.", nil, nil)
	pgxWaitDesc = prometheus.NewDesc(
		"app_db_pool_acquire_wait_seconds_total",
		"Cumulative time callers spent waiting for a connection.", nil, nil)
	pgxEmptyAcquireDesc = prometheus.NewDesc(
		"app_db_pool_empty_acquires_total",
		"Acquires that had to wait because the pool was empty.", nil, nil)
)

func (c *pgxPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- pgxTotalDesc
	ch <- pgxIdleDesc
	ch <- pgxAcquiredDesc
	ch <- pgxMaxDesc
	ch <- pgxWaitDesc
	ch <- pgxEmptyAcquireDesc
}

func (c *pgxPoolCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.pool.Stat()
	ch <- prometheus.MustNewConstMetric(pgxTotalDesc, prometheus.GaugeValue, float64(s.TotalConns()))
	ch <- prometheus.MustNewConstMetric(pgxIdleDesc, prometheus.GaugeValue, float64(s.IdleConns()))
	ch <- prometheus.MustNewConstMetric(pgxAcquiredDesc, prometheus.GaugeValue, float64(s.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(pgxMaxDesc, prometheus.GaugeValue, float64(s.MaxConns()))
	ch <- prometheus.MustNewConstMetric(pgxWaitDesc, prometheus.CounterValue, s.AcquireDuration().Seconds())
	ch <- prometheus.MustNewConstMetric(pgxEmptyAcquireDesc, prometheus.CounterValue, float64(s.EmptyAcquireCount()))
}

type redisPoolCollector struct {
	client *redis.Client
}

var (
	redisHitsDesc = prometheus.NewDesc(
		"app_redis_pool_hits_total", "Connections served from the Redis pool.", nil, nil)
	redisMissesDesc = prometheus.NewDesc(
		"app_redis_pool_misses_total", "Connections the Redis pool had to open.", nil, nil)
	redisTimeoutsDesc = prometheus.NewDesc(
		"app_redis_pool_timeouts_total", "Times a caller timed out waiting for a Redis connection.", nil, nil)
	redisTotalDesc = prometheus.NewDesc(
		"app_redis_pool_total_conns", "Total connections in the Redis pool.", nil, nil)
	redisIdleDesc = prometheus.NewDesc(
		"app_redis_pool_idle_conns", "Idle connections in the Redis pool.", nil, nil)
)

func (c *redisPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- redisHitsDesc
	ch <- redisMissesDesc
	ch <- redisTimeoutsDesc
	ch <- redisTotalDesc
	ch <- redisIdleDesc
}

func (c *redisPoolCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.client.PoolStats()
	ch <- prometheus.MustNewConstMetric(redisHitsDesc, prometheus.CounterValue, float64(s.Hits))
	ch <- prometheus.MustNewConstMetric(redisMissesDesc, prometheus.CounterValue, float64(s.Misses))
	ch <- prometheus.MustNewConstMetric(redisTimeoutsDesc, prometheus.CounterValue, float64(s.Timeouts))
	ch <- prometheus.MustNewConstMetric(redisTotalDesc, prometheus.GaugeValue, float64(s.TotalConns))
	ch <- prometheus.MustNewConstMetric(redisIdleDesc, prometheus.GaugeValue, float64(s.IdleConns))
}
