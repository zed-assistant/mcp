package localidp

import "github.com/zed-assistant/mcp/internal/configuration"

type LocalIDP struct {
	appConfig *configuration.AppConfig
}

func NewLocalIDP(appConfig *configuration.AppConfig) *LocalIDP {
	return &LocalIDP{
		appConfig: appConfig,
	}
}

func (l *LocalIDP) GetAuthorizationURL(state string, nonce string) (string, error) {
	url := l.appConfig.Server.ExternalUrl + "/auth/local" + "?state=" + state + "&nonce=" + nonce
	return url, nil
}
