package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"

	"gitlab.unanet.io/devops/go/pkg/metrics"
)

// Metrics adapts the incoming request with Logging/Metrics
func Metrics(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Tally the incoming request metrics
		now := time.Now()
		metrics.StatHTTPRequestCount.WithLabelValues(r.Method, r.Proto, r.URL.Path).Inc()
		metrics.StatRequestSaturationGauge.WithLabelValues(r.Method, r.Proto, r.URL.Path).Inc()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Run this on the way out (i.e. outgoing response)
		defer func() {
			// Calculate the request duration (i.e. latency)
			ms := float64(time.Since(now).Milliseconds())

			// Tally the outgoing response metrics
			metrics.StatRequestDurationHistogram.WithLabelValues(r.Method, r.Proto, r.URL.Path).Observe(ms)
			metrics.StatRequestDurationGauge.WithLabelValues(r.Method, r.Proto, r.URL.Path).Set(ms)
			metrics.StatRequestSaturationGauge.WithLabelValues(r.Method, r.Proto, r.URL.Path).Dec()
			metrics.StatHTTPResponseCount.WithLabelValues(strconv.Itoa(ww.Status()), r.Method, r.Proto, r.URL.Path).Inc()
		}()
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
