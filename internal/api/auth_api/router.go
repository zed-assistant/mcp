package authapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ory/fosite"
	httpmiddleware "github.com/zed-assistant/mcp/internal/api/http_middleware"
	"github.com/zed-assistant/mcp/internal/auth/idp"
	"github.com/zed-assistant/mcp/internal/auth/oauth"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/logger"
)

type IDPManger interface {
	VerifyAssertion(assertion string) (*idp.AuthenticationResult, error)
}
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
	idpManager       IDPManger
	localIDP         LocalIDP
	getCurrentTime   func() time.Time
}

func NewAuthApi(
	appConfig *configuration.AppConfig,
	oauthProvider fosite.OAuth2Provider,
	oauthStore *oauth.MemoryStore,
	pendingAuthStore *oauth.PendingStore,
	log *slog.Logger,
	idpManager IDPManger,
	localIDP LocalIDP,
	getCurrentTime func() time.Time,
) *AuthApi {
	return &AuthApi{
		appConfig:        appConfig,
		oauthProvider:    oauthProvider,
		oauthStore:       oauthStore,
		pendingAuthStore: pendingAuthStore,
		log:              log,
		idpManager:       idpManager,
		localIDP:         localIDP,
		getCurrentTime:   getCurrentTime,
	}
}

func (a *AuthApi) GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(httpmiddleware.AnonymousCORSMiddleware())

	router.Get("/authorize", a.authorize)
	router.Get("/local", a.localAuthentication)
	router.Get("/callback", a.authenticationCallback)
	router.Post("/token", a.getToken)
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
