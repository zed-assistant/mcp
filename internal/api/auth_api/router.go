package authapi

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/ory/fosite"
	"github.com/zed-assistant/mcp/internal/auth/oauth"
)

type IDPManager interface {
	GetAuthorizationURL(state string, nonce string) (string, error)
}

type AuthApi struct {
	oauthProvider    fosite.OAuth2Provider
	oauthStore       *oauth.MemoryStore
	pendingAuthStore *oauth.PendingStore
	log              *slog.Logger
	idpManager       IDPManager
}

func NewAuthApi(oauthProvider fosite.OAuth2Provider, oauthStore *oauth.MemoryStore, pendingAuthStore *oauth.PendingStore, log *slog.Logger, idpManager IDPManager) *AuthApi {
	return &AuthApi{
		oauthProvider:    oauthProvider,
		oauthStore:       oauthStore,
		pendingAuthStore: pendingAuthStore,
		log:              log,
		idpManager:       idpManager,
	}
}

func (a *AuthApi) GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/authorize", a.authorize)
	return router
}
