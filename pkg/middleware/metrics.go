package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.unanet.io/devops/go/pkg/metrics"
)

// Metrics adapts the incoming request with Logging/Metrics
func Metrics(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Incoming Request TimeStamp
		now := time.Now()
		// Wrapped Response (so we can get the status code on the way out)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		// Call the next handler and then tally the metrics
		// https://github.com/go-chi/chi/blob/master/context.go
		next.ServeHTTP(ww, r)

		chiRouteCtx := chi.RouteContext(r.Context())
		var routePattern string
		if chiRouteCtx != nil {
			routePattern = chiRouteCtx.RoutePattern()
		}
		metrics.StatHTTPRequestCount.WithLabelValues(r.Method, r.Proto, routePattern).Inc()
		metrics.StatRequestSaturationGauge.WithLabelValues(r.Method, r.Proto, routePattern).Inc()
		// Run this on the way out (i.e. outgoing response)
		defer func() {
			// Calculate the request duration (i.e. latency)
			seconds := time.Since(now).Seconds()

			// Tally the outgoing response metrics
			metrics.StatRequestDurationHistogram.WithLabelValues(r.Method, r.Proto, routePattern).Observe(seconds)
			metrics.StatRequestSaturationGauge.WithLabelValues(r.Method, r.Proto, routePattern).Dec()
			metrics.StatHTTPResponseCount.WithLabelValues(strconv.Itoa(ww.Status()), r.Method, r.Proto, routePattern).Inc()
		}()

	}
	return http.HandlerFunc(fn)
}
