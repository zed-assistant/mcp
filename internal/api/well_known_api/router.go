package wellknownapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpmiddleware "github.com/zed-assistant/mcp/internal/api/http_middleware"
	"github.com/zed-assistant/mcp/internal/configuration"
)

type WellKnownApi struct {
	appConfig *configuration.AppConfig
}

func NewWellKnownApi(appConfig *configuration.AppConfig) *WellKnownApi {
	return &WellKnownApi{
		appConfig: appConfig,
	}
}

func (a *WellKnownApi) GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(httpmiddleware.AnonymousCORSMiddleware())

	router.Get("/oauth-authorization-server", a.getAuthServerdMetadata)
	router.Get("/oauth-protected-resource/mcp", a.getMCPProtectedResourceMetadata)

	return router
}

func writeWellKnownJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_ = json.NewEncoder(w).Encode(v)
}
