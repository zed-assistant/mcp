package authapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ory/fosite"
	"github.com/zed-assistant/mcp/internal/auth/oauth"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/logger"
)

type LocalIDP interface {
	GetAuthorizationURL(state string, nonce string) (string, error)
	Authenticate(username string, password string, pendingRequestID string) (string, error)
}

type AuthApi struct {
	appConfig        *configuration.AppConfig
	oauthProvider    fosite.OAuth2Provider
	oauthStore       *oauth.MemoryStore
	pendingAuthStore *oauth.PendingStore
	log              *slog.Logger
	localIDP         LocalIDP
}

func NewAuthApi(
	appConfig *configuration.AppConfig,
	oauthProvider fosite.OAuth2Provider,
	oauthStore *oauth.MemoryStore,
	pendingAuthStore *oauth.PendingStore,
	log *slog.Logger,
	localIDP LocalIDP,
) *AuthApi {
	return &AuthApi{
		appConfig:        appConfig,
		oauthProvider:    oauthProvider,
		oauthStore:       oauthStore,
		pendingAuthStore: pendingAuthStore,
		log:              log,
		localIDP:         localIDP,
	}
}

func (a *AuthApi) GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/authorize", a.authorize)
	router.Get("/local", a.localAuthentication)
	return router
}

func (a *AuthApi) writeHTMLResponse(ctx context.Context, w http.ResponseWriter, statusCode int, htmlContent string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(htmlContent))
	if err != nil {
		a.log.ErrorContext(ctx, "Error writing response", logger.LogError(err))
	}
}
