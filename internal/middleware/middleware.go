// internal/middleware/middleware.go

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/panditvishnuu/userservice/internal/contextkeys"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Inject into context so handlers and downstream middleware can read it.
		ctx := context.WithValue(r.Context(), contextkeys.RequestIDKey, requestID)

		// Inject into response so the caller can correlate their request.
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
// The standard ResponseWriter provides no way to read back what was written.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
