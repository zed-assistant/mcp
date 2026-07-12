package authorization

import (
	"github.com/ory/fosite"
	"github.com/zed-assistant/mcp/internal/configuration"
)

type AuthorizationManager struct {
	appConfig     *configuration.AppConfig
	oauthProvider fosite.OAuth2Provider
}

func NewAuthorizationManager(appConfig *configuration.AppConfig, oauthProvider fosite.OAuth2Provider) *AuthorizationManager {
	return &AuthorizationManager{
		appConfig:     appConfig,
		oauthProvider: oauthProvider,
	}
}
