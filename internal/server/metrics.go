package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	grpcRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total gRPC requests by method and status code.",
	}, []string{"method", "code"})

	grpcDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "gRPC request latency in seconds by method.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})
)

func recordMetrics(method, code string, dur time.Duration) {
	grpcRequests.WithLabelValues(method, code).Inc()
	grpcDuration.WithLabelValues(method).Observe(dur.Seconds())
}

// MetricsHandler serves this service's Prometheus metrics.
func MetricsHandler() http.Handler { return promhttp.Handler() }
