package idp

import (
	"fmt"
	"net/url"

	"github.com/zed-assistant/mcp/internal/configuration"
)

type IDPManager struct {
	appConfig *configuration.AppConfig
}

func NewIDPManager(appConfig *configuration.AppConfig) *IDPManager {
	return &IDPManager{
		appConfig: appConfig,
	}
}

func (m *IDPManager) GetAuthorizationURL(state string, nonce string) (string, error) {
	switch m.appConfig.OAuth2.IDP.Type {
	case "local":
		return appendQueryParams(m.appConfig.Server.ExternalUrl+"/auth/local", state, nonce)
	default:
		return "", fmt.Errorf("unsupported IDP type: %s", m.appConfig.OAuth2.IDP.Type)
	}
}

func appendQueryParams(baseURL string, state string, nonce string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %v", err)
	}

	q := u.Query()
	q.Set("state", state)
	q.Set("nonce", nonce)
	u.RawQuery = q.Encode()

	return u.String(), nil
}