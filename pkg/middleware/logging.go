package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/unanet/go/v2/pkg/log"
)

type LogWriter interface {
	Write(status, bytes int, header http.Header, elapsed time.Duration, body io.ReadCloser)
	Panic(v interface{}, stack []byte)
}

type LogConstructor interface {
	NewLogWriter(r *http.Request) LogWriter
}

type LogEntry struct {
	logger *zap.Logger
}

func WithLogEntry(r *http.Request, entry LogWriter) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), middleware.LogEntryCtxKey, entry))
	return r
}

func RequestLogger(f LogConstructor) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := f.NewLogWriter(r)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			w.Header().Add(log.RequestIDHeader, log.GetReqID(r.Context()))
			t1 := time.Now()
			buffer := &limitedRW{limit: 1024 * 512}
			if log.Logger.Core().Enabled(zap.DebugLevel) {
				ww.Tee(buffer)
			}
			defer func() {
				entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), io.NopCloser(buffer))
			}()

			next.ServeHTTP(ww, WithLogEntry(r, entry))
		}
		return http.HandlerFunc(fn)
	}
}

func (l *LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, body io.ReadCloser) {
	outgoingResponseFields := []zap.Field{
		zap.Int("status", status),
		zap.Int("resp_bytes_length", bytes),
		zap.Float64("elapsed", elapsed.Seconds()),
	}

	if l.logger.Core().Enabled(zap.DebugLevel) {
		outgoingResponseFields = append(outgoingResponseFields, log.DecodeHeader(header))
		outgoingResponseFields = append(outgoingResponseFields, log.DecodeBody(body))
	}

	l.logger.With(outgoingResponseFields...).Info("Outgoing HTTP Response")
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	l.logger.Panic(fmt.Sprintf("%+v", v),
		zap.String("stack", string(stack)))
}

func Logger() func(next http.Handler) http.Handler {
	return RequestLogger(&LogWriterConstructor{log.Logger})
}

type LogWriterConstructor struct {
	logger *zap.Logger
}

func (l *LogWriterConstructor) NewLogWriter(r *http.Request) LogWriter {
	logFields := []zap.Field{
		zap.String("user_agent", r.UserAgent()),
	}

	incomingRequestFields := []zap.Field{
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("uri", r.RequestURI),
		zap.String("method", r.Method),
	}

	if l.logger.Core().Enabled(zap.DebugLevel) {
		incomingRequestFields = append(incomingRequestFields, log.DecodeBodyFromRequest(r))
		incomingRequestFields = append(incomingRequestFields, log.DecodeHeaderFromRequest(r))
	}

	if reqID := log.GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, zap.String("req_id", reqID))
	}

	entry := &LogEntry{
		logger: l.logger.With(logFields...),
	}

	entry.logger.With(incomingRequestFields...).Info("Incoming HTTP Request")

	return entry
}

func LogFromRequest(r *http.Request) *zap.Logger {
	return Log(r.Context())
}

func Log(ctx context.Context) *zap.Logger {
	v := ctx.Value(middleware.LogEntryCtxKey)
	if v == nil {
		return log.Logger
	}
	if entry, ok := v.(*LogEntry); ok {
		return entry.logger
	} else {
		return log.Logger
	}
}

type limitedRW struct {
	size  int
	limit int
	buf   bytes.Buffer
}

func (l *limitedRW) Write(b []byte) (int, error) {
	if l.size >= l.limit {
		return len(b), nil
	}

	n, err := l.buf.Write(b)
	l.size += n
	return n, err
}

func (l *limitedRW) Read(b []byte) (int, error) {
	return l.buf.Read(b)
}
