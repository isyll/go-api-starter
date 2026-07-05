package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "grpc",
			Name:      "requests_total",
			Help:      "Total unary RPCs handled, by full method and status code.",
		},
		[]string{"method", "code"},
	)

	GRPCRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "grpc",
			Name:      "request_duration_seconds",
			Help:      "Unary RPC latency, by full method.",
			Buckets: []float64{
				0.001, 0.0025, 0.005, 0.01, 0.025, 0.05,
				0.1, 0.25, 0.5, 1, 2.5, 5, 10,
			},
		},
		[]string{"method"},
	)

	GRPCRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "grpc",
			Name:      "requests_in_flight",
			Help:      "Unary RPCs currently being handled.",
		},
	)

	BuildInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "build_info",
			Help:      "Build metadata; the value is always 1.",
		},
		[]string{"version", "commit"},
	)
)
