package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	unaryLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_unary_latency_seconds",
			Help:    "Latency of unary RPCs",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)
	streamLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_stream_latency_seconds",
			Help:    "Latency of streaming RPCs",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)
)

func init() {
	prometheus.MustRegister(unaryLatency, streamLatency)
}

func observeUnary(method, code string, d time.Duration) {
	unaryLatency.WithLabelValues(method, code).Observe(d.Seconds())
}

func observeStream(method, code string, d time.Duration) {
	streamLatency.WithLabelValues(method, code).Observe(d.Seconds())
}

func startMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		_ = http.ListenAndServe(":9090", mux)
	}()
}
