package instrumentation

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

// Initialize Prometheus metrics
func initPrometheus() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(SuccessRequestCounter)
	prometheus.MustRegister(FailureRequestCounter)
	prometheus.MustRegister(SuccessLatency)
	prometheus.MustRegister(FailureLatency)
}

// Start Prometheus metrics server
func StartPrometheusServer() {
	initPrometheus()
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Prometheus server running at :%s/metrics\n", viper.GetString("prometheus.port"))
	go http.ListenAndServe(":"+viper.GetString("prometheus.port"), nil)
}
