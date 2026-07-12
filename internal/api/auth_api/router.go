package authapi

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/ory/fosite"
	"github.com/zed-assistant/mcp/internal/auth/oauth"
)

type AuthApi struct {
	oauthProvider fosite.OAuth2Provider
	oauthStore    *oauth.MemoryStore
	log           *slog.Logger
}

func NewAuthApi(oauthProvider fosite.OAuth2Provider, oauthStore *oauth.MemoryStore, log *slog.Logger) *AuthApi {
	return &AuthApi{
		oauthProvider: oauthProvider,
		oauthStore:    oauthStore,
		log:           log,
	}
}

func (a *AuthApi) GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/authorize", a.authorize)
	return router
}
