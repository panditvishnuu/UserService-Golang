package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/panditvishnuu/userservice/internal/contextkeys"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// We need to capture the status code written by the handler.
		// http.ResponseWriter doesn't expose it after the fact, so we wrap it.
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		requestID, _ := r.Context().Value(contextkeys.RequestIDKey).(string)

		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestID,
		)
	})
}
