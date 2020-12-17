package middleware

import (
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func SetupMiddleware(r chi.Router, timeout time.Duration) {
	r.Use(Metrics)
	r.Use(RequestID)
	r.Use(middleware.RealIP)
	r.Use(Logger())
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(timeout))
}

type MetricsProvider interface {
}
