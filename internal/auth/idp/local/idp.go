package localidp

import (
	"time"

	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/jwt"
)

type LocalIDP struct {
	appConfig      *configuration.AppConfig
	signJwt        func(claims jwt.Claims, options jwt.SigningOptions) (string, error)
	getCurrentTime func() time.Time
}

func NewLocalIDP(
	appConfig *configuration.AppConfig,
	signJwt func(claims jwt.Claims, options jwt.SigningOptions) (string, error),
	getCurrentTime func() time.Time,
) *LocalIDP {
	return &LocalIDP{
		appConfig:      appConfig,
		signJwt:        signJwt,
		getCurrentTime: getCurrentTime,
	}
}

func (l *LocalIDP) GetAuthorizationURL(state string, nonce string) (string, error) {
	url := l.appConfig.Server.ExternalUrl + "/auth/local" + "?state=" + state + "&nonce=" + nonce
	return url, nil
}
