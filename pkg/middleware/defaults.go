package middleware

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupMiddleware(r chi.Router, timeout time.Duration) {
	r.Use(Metrics)
	r.Use(RequestID)
	r.Use(Messaging)
	r.Use(middleware.RealIP)
	r.Use(Logger())
	r.Use(middleware.Recoverer)
	if timeout > 0 {
		r.Use(middleware.Timeout(timeout))
	}
}

type MetricsProvider interface {
}
