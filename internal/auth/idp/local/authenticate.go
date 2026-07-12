package localidp

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/zed-assistant/mcp/internal/auth/idp"
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

		authResult := &idp.AuthenticationResult{
			Email:            user.Username,
			Sub:              sub,
			IDP:              "local",
			PendingRequestID: pendingRequestID,
		}

		signed, err := i.idpManger.SignAssertion(authResult)
		if err != nil {
			return "", fmt.Errorf("failed to sign assertion: %w", err)
		}

		return signed, nil
	}

	return "", ErrUserNotFound
}
