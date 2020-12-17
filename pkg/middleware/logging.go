package middleware

import (
"context"
"fmt"
	"net/http"
"time"

"gitlab.unanet.io/devops/go/pkg/log"

"github.com/go-chi/chi/middleware"

"go.uber.org/zap"
)

type LogEntry struct {
	logger *zap.Logger
}

func (l *LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.logger.Info("Outgoing HTTP Response",
		zap.Int("status", status),
		zap.Int("resp_bytes_length", bytes),
		zap.Float64("elapsed_ms", float64(elapsed.Nanoseconds())/1000000.0))
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	l.logger.Panic(fmt.Sprintf("%+v", v),
		zap.String("stack", string(stack)))
}

func Logger() func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&LogEntryConstructor{log.Logger})
}

type LogEntryConstructor struct {
	logger *zap.Logger
}

func (l *LogEntryConstructor) NewLogEntry(r *http.Request) middleware.LogEntry {
	logFields := []zap.Field{
		zap.String("user_agent", r.UserAgent()),
	}

	incomingRequestFields := []zap.Field{
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("uri", r.RequestURI),
		zap.String("method", r.Method),
	}

	if reqID := GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, zap.String("req_id", reqID))
		incomingRequestFields = append(incomingRequestFields, zap.String("req_id", reqID))
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
