package retry

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type retrier struct {
	backoff  float32
	interval time.Duration
	logger   logger
}

type logger interface {
	Printf(format string, v ...interface{})
}

type noopLogger struct{}

func (n noopLogger) Printf(format string, v ...interface{}) {}

type Opt func(*retrier)

func WithLogger(l logger) Opt {
	return func(r *retrier) {
		r.logger = l
	}
}

func WithInterval(b float32) Opt {
	return func(r *retrier) {
		r.backoff = b
	}
}

func Do(ctx context.Context, fn func() error, opts ...Opt) error {
	r := retrier{
		backoff:  1.4,
		interval: time.Millisecond * 200,
		logger:   noopLogger{},
	}
	for _, o := range opts {
		o(&r)
	}

	t := NewDecayTimer(r.interval, r.backoff)
	defer t.Stop()

	var err error
	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(err, "return context timeout exceeded")
		case <-t.C:
			err = fn()
			if err == nil {
				return nil
			}
			r.logger.Printf("retrier encountered an error: %v", err)
		}
	}
}
