package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

// newLoggingRW ...
func newLoggingRW(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (l *responseWriter) Status() int {
	return l.statusCode
}

// WriteHeader changing internal field and writes code to header.
func (l *responseWriter) WriteHeader(statusCode int) {
	if l.wroteHeader {
		return
	}
	l.statusCode = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
	l.wroteHeader = true
}

// LogRequest return middleware function that will log meta info about every request to logger used in initializing mw.
func LogRequest(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			id := middleware.GetReqID(r.Context())
			lrw := newLoggingRW(w)

			defer func() {
				if rcv := recover(); rcv != nil {
					lrw.WriteHeader(http.StatusInternalServerError)
					logger.Error("recovered panic", zap.Any("recovered", rcv), zap.ByteString("debug stack", debug.Stack()))
				}
			}()

			// start check time
			start := time.Now()
			next.ServeHTTP(w, r)
			dur := time.Since(start)

			var level zapcore.Level
			switch {
			case lrw.statusCode >= 500:
				level = zap.ErrorLevel
			case lrw.statusCode >= 400:
				level = zap.InfoLevel
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
