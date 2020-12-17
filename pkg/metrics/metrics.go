package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"gitlab.unanet.io/devops/go/pkg/log"
)

var (
	StatRequestSaturationGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_saturation",
			Help: "The total number of requests inside the server (transactions serving)",
		}, []string{"method", "protocol", "path"})

	StatBuildInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_build_info",
			Help: "A metric with a constant '1' value labeled by version, revision, branch, and goversion from which the service was built",
		}, []string{"service", "revision", "branch", "version", "author", "build_date", "build_user", "build_host"})

	StatRequestDurationGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_duration_ms",
			Help: "The time the server spends processing a request in milliseconds",
		}, []string{"method", "protocol", "path"})

	StatHTTPRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "The total number of incoming requests to the service",
		}, []string{"method", "protocol", "path"})

	StatHTTPResponseCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_total",
			Help: "The total number of outgoing responses to the client",
		}, []string{"code", "method", "protocol", "path"})

	StatRequestDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_histogram_ms",
			Help:    "time spent processing an http request in milliseconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 18),
		}, []string{"method", "protocol", "path"})
)

func StartMetricsServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		ReadTimeout:  time.Duration(5) * time.Second,
		WriteTimeout: time.Duration(30) * time.Second,
		IdleTimeout:  time.Duration(90) * time.Second,
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Panic("Failed to Start Metrics Server", zap.Error(err))
		}

		log.Logger.Info("Metrics Server Shutdown")
	}()

	log.Logger.Info("Metrics Listener", zap.Int("port", port))
	return server
}
