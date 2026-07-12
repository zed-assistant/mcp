package httpmiddleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/zed-assistant/mcp/internal/appctx"
)

var requestIdHeader = "X-Request-Id"

func RequestIdMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := uuid.New()

			ctx := r.Context()

			w.Header().Set("X-Request-Id", requestId.String())

			ctx = appctx.WithRequestId(ctx, requestId.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
