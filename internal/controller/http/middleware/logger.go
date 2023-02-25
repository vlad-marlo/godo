package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggingRW struct {
	http.ResponseWriter
	statusCode int
}

// newLoggingRW ...
func newLoggingRW(w http.ResponseWriter) *loggingRW {
	return &loggingRW{w, http.StatusOK}
}

// WriteHeader changing internal field and writes code to header.
func (l *loggingRW) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

// LogRequest return middleware function that will log meta info about every request to logger used in initializing mw.
func LogRequest(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := middleware.GetReqID(r.Context())
			lrw := newLoggingRW(w)

			// start check time
			start := time.Now()
			next.ServeHTTP(lrw, r)
			dur := time.Since(start)

			var level zapcore.Level
			switch {
			case lrw.statusCode >= 500:
				level = zap.ErrorLevel
			case lrw.statusCode >= 400:
				level = zap.DebugLevel
			default:
				level = zap.DebugLevel
			}
			// log request
			logger.Log(
				level,
				fmt.Sprintf("status text: %s", http.StatusText(lrw.statusCode)),
				zap.String("method", r.Method),
				zap.String("url", r.URL.Path),
				zap.Duration("duration", dur),
				zap.Int("code", lrw.statusCode),
				zap.String("request_id", id),
			)
		})
	}
}
