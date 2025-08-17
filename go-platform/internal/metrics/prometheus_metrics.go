package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	providerQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "detectviz_prometheus_provider_queries_total",
			Help: "Total number of queries processed by the Prometheus provider.",
		},
		[]string{"status"}, // e.g., success, error, cache_hit, circuit_breaker_rejected
	)

	providerQueryDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "detectviz_prometheus_provider_query_duration_seconds",
			Help:    "Histogram of the latency of queries to the backend Prometheus server.",
			Buckets: prometheus.DefBuckets,
		},
	)

	providerCacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "detectviz_prometheus_provider_cache_hits_total",
			Help: "Total number of cache hits.",
		},
	)

	providerCacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "detectviz_prometheus_provider_cache_misses_total",
			Help: "Total number of cache misses.",
		},
	)

	providerCircuitBreakerState = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "detectviz_prometheus_provider_circuit_breaker_state",
			Help: "The current state of the circuit breaker (0: closed, 1: open, 2: half-open).",
		},
	)

	providerInflightQueries = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "detectviz_prometheus_provider_inflight_queries",
			Help: "Number of in-flight queries to the backend Prometheus server.",
		},
	)
)
