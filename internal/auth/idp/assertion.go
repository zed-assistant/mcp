package idp

import (
	"fmt"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/zed-assistant/mcp/internal/jwt"
)

type AuthenticationResult struct {
	Email            string
	Sub              string
	IDP              string
	PendingRequestID string
}

func (m *IDPManger) SignAssertion(authResult *AuthenticationResult) (string, error) {
	now := m.getCurrentTime()

	claims := jwt.Claims{
		Issuer:    "local",
		Subject:   authResult.Sub,
		Audience:  []string{m.appConfig.Server.ExternalUrl},
		Expiry:    now.Add(5 * time.Minute),
		IssuedAt:  now,
		NotBefore: now,
		Additional: map[string]any{
			"AuthResult": authResult,
		},
	}

	signingOpts := jwt.SigningOptions{
		Secret: m.signingSecret,
	}

	signed, err := m.signJwt(claims, signingOpts)
	if err != nil {
		return "", fmt.Errorf("failed to sign assertion: %w", err)
	}
	return signed, nil
}

func (m *IDPManger) VerifyAssertion(assertion string) (*AuthenticationResult, error) {
	claims, err := m.verifyJwt(assertion, jwt.VerifyingOptions{
		Secret:           m.signingSecret,
		ExpectedAudience: m.appConfig.Server.ExternalUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify assertion: %w", err)
	}

	authResult, ok := claims.Additional["AuthResult"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse authentication result as map from assertion")
	}

	var result AuthenticationResult
	err = mapstructure.Decode(authResult, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode authentication result from assertion: %w", err)
	}

	return &result, nil
}
