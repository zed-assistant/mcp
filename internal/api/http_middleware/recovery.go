package httpmiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap the response writer to track if headers were written
			rw := &responseWriter{ResponseWriter: w}

			defer func(ctx context.Context) {
				if rvr := recover(); rvr != nil {
					// Don't recover from http.ErrAbortHandler
					if err, ok := rvr.(error); ok && errors.Is(err, http.ErrAbortHandler) {
						panic(rvr)
					}

					logger.ErrorContext(ctx, "Panic recovered",
						slog.Any("error", rvr),
						slog.String("stack", string(debug.Stack())))

					// Check if headers were already written
					if rw.written {
						// Headers already sent, cannot send error response
						return
					}

					// Only send error response for non-websocket connections
					if r.Header.Get("Connection") != "Upgrade" {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						_ = json.NewEncoder(w).Encode(map[string]string{
							"message": "Unexpected server error",
						})
					}
				}
			}(r.Context())

			next.ServeHTTP(rw, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to track if headers were written
type responseWriter struct {
	http.ResponseWriter

	written bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.written = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.written = true
	return rw.ResponseWriter.Write(data)
}
