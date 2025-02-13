package instrumentation

import (
	"fortis/entity/constants"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometheus CounterVec for tracking the number of HTTP requests.
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of http requests",
		},
		[]string{constants.Handler},
	)
	SuccessRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_success_request_total",
			Help: "Total number of successful http requests",
		},
		[]string{constants.Handler},
	)
	FailureRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_failed_request_total",
			Help: "Total number of failed http requests",
		},
		[]string{constants.Handler, constants.Status},
	)

	// Prometheus HistogramVec for tracking the latency of HTTP requests in seconds.
	SuccessLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "successful_http_request_latency_seconds",
			Help:    "Latency of the successful HTTP requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{constants.Handler},
	)
	FailureLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "failed_http_request_latency_seconds",
			Help:    "Latency of the failed HTTP requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{constants.Handler, constants.Status},
	)
)
