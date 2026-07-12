package oauth

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/zed-assistant/mcp/internal/configuration"
)

type Random interface {
	RandomBytes(length int) ([]byte, error)
	RandomBytesHex(length int) (string, error)
}

type Store interface {
	fosite.ClientManager
	oauth2.CoreStorage
	oauth2.TokenRevocationStorage
}

func NewOAuth2Provider(appConfig *configuration.AppConfig, random Random, store Store) (fosite.OAuth2Provider, error) {
	globalSecret := appConfig.OAuth2.SigningSecret
	if len(globalSecret) == 0 {
		s, err := random.RandomBytes(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random signing secret: %w", err)
		}
		globalSecret = s
	}

	idTokenSigningKey := appConfig.OAuth2.IdTokenSigningKey
	if idTokenSigningKey == nil {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID token signing key: %w", err)
		}
		idTokenSigningKey = (*configuration.RSAPrivateKey)(k)
	}

	cfg := &fosite.Config{
		GlobalSecret:          globalSecret,
		AccessTokenLifespan:   time.Hour,
		RefreshTokenLifespan:  30 * 24 * time.Hour,
		AuthorizeCodeLifespan: time.Minute,

		EnforcePKCE:                    true,
		EnablePKCEPlainChallengeMethod: false,

		SendDebugMessagesToClients: false,
	}

	return compose.ComposeAllEnabled(cfg, store, idTokenSigningKey), nil
}
