package wellknownapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpmiddleware "github.com/zed-assistant/mcp/internal/api/http_middleware"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/logger"
)

type WellKnownApi struct {
	appConfig *configuration.AppConfig
	log       *slog.Logger
}

func NewWellKnownApi(appConfig *configuration.AppConfig, log *slog.Logger) *WellKnownApi {
	return &WellKnownApi{
		appConfig: appConfig,
		log:       log,
	}
}

func (a *WellKnownApi) GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(httpmiddleware.AnonymousCORSMiddleware())

	router.Get("/oauth-authorization-server", a.getAuthServerdMetadata)
	router.Get("/oauth-protected-resource/mcp", a.getMCPProtectedResourceMetadata)

	return router
}

func (a *WellKnownApi) writeWellKnownJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		a.log.Error("Error writing well-known JSON response", logger.LogError(err))
	}
}
