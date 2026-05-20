package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/logger"
)

// statusRecorder wraps http.ResponseWriter to capture the HTTP status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// LoggingMiddleware logs every HTTP request at TRACE (start + end) and then
// at the appropriate level based on method and response status:
//
//   - DEBUG  — successful read-only requests (GET/HEAD 2xx)
//   - INFO   — successful mutating requests (POST/PUT/DELETE 2xx)
//   - WARN   — client errors (4xx)
//   - ERROR  — server errors (5xx)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		logger.Trace(ctx, "request started", "method", r.Method, "path", r.URL.Path)

		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		duration := time.Since(start)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration", duration,
		}

		logger.Trace(ctx, "request completed", attrs...)

		switch {
		case status >= 500:
			slog.ErrorContext(ctx, "request failed", attrs...)
		case status >= 400:
			slog.WarnContext(ctx, "request error", attrs...)
		case r.Method == http.MethodGet || r.Method == http.MethodHead:
			slog.DebugContext(ctx, "request completed", attrs...)
		default:
			slog.InfoContext(ctx, "request completed", attrs...)
		}
	})
}
