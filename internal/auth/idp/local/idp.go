package localidp

import (
	"github.com/zed-assistant/mcp/internal/auth/idp"
	"github.com/zed-assistant/mcp/internal/configuration"
)

type IDPManger interface {
	SignAssertion(authResult *idp.AuthenticationResult) (string, error)
}

type LocalIDP struct {
	appConfig *configuration.AppConfig
	idpManger IDPManger
}

func NewLocalIDP(
	appConfig *configuration.AppConfig,
	idpManger IDPManger,
) *LocalIDP {
	return &LocalIDP{
		appConfig: appConfig,
		idpManger: idpManger,
	}
}

func (l *LocalIDP) GetAuthorizationURL(state string, nonce string) (string, error) {
	url := l.appConfig.Server.ExternalUrl + "/auth/local" + "?state=" + state + "&nonce=" + nonce
	return url, nil
}
