package idp

import (
	"time"

	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/jwt"
	"github.com/zed-assistant/mcp/internal/random"
)

type IDPManger struct {
	appConfig      *configuration.AppConfig
	signJwt        func(claims jwt.Claims, options jwt.SigningOptions) (string, error)
	verifyJwt      func(token string, options jwt.VerifyingOptions) (*jwt.Claims, error)
	getCurrentTime func() time.Time
	signingSecret  []byte
}

func NewIDPManger(
	appConfig *configuration.AppConfig,
	signJwt func(claims jwt.Claims, options jwt.SigningOptions) (string, error),
	verifyJwt func(token string, options jwt.VerifyingOptions) (*jwt.Claims, error),
	getCurrentTime func() time.Time,
	random *random.Random,
) (*IDPManger, error) {
	signingSecret := []byte(appConfig.OAuth2.SigningSecret)
	if len(signingSecret) == 0 {
		s, err := random.RandomBytes(32)
		if err != nil {
			return nil, err
		}
		signingSecret = s
	}

	return &IDPManger{
		appConfig:      appConfig,
		signJwt:        signJwt,
		verifyJwt:      verifyJwt,
		getCurrentTime: getCurrentTime,
		signingSecret:  signingSecret,
	}, nil
}
