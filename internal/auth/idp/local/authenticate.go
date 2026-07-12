package localidp

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/zed-assistant/mcp/internal/auth/idp"
	"github.com/zed-assistant/mcp/internal/jwt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserNotFound = errors.New("user not found")

func (i *LocalIDP) Authenticate(username string, password string, pendingRequestID string) (string, error) {
	for _, user := range i.appConfig.OAuth2.IDP.Local.Users {
		if user.Username != username {
			continue
		}
		if user.Password != password {
			return "", ErrInvalidCredentials
		}

		usernameHash := sha256.Sum256([]byte(user.Username))
		sub := hex.EncodeToString(usernameHash[:])

		now := i.getCurrentTime()

		claims := jwt.Claims{
			Issuer:    "local",
			Subject:   sub,
			Audience:  []string{i.appConfig.Server.ExternalUrl},
			Expiry:    now.Add(5 * time.Minute),
			IssuedAt:  now,
			NotBefore: now,
			Additional: map[string]any{
				"AuthResult": idp.AuthenticationResult{
					Email:            user.Username,
					Sub:              sub,
					IDP:              "local",
					PendingRequestID: pendingRequestID,
				},
			},
		}

		signingOpts := jwt.SigningOptions{}

		signed, err := i.signJwt(claims, signingOpts)
		if err != nil {
			return "", fmt.Errorf("failed to sign jwt: %w", err)
		}

		return signed, nil
	}

	return "", ErrUserNotFound
}
