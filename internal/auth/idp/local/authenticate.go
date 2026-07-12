package localidp

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type LocalIDPUser struct {
	Email string
	Sub   string
}

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserNotFound = errors.New("user not found")

func (i *LocalIDP) Authenticate(username string, password string) (*LocalIDPUser, error) {
	for _, user := range i.appConfig.OAuth2.IDP.Local.Users {
		if user.Username != username {
			continue
		}
		if user.Password != password {
			return nil, ErrInvalidCredentials
		}

		usernameHash := sha256.Sum256([]byte(user.Username))
		return &LocalIDPUser{
			Email: user.Username,
			Sub:   hex.EncodeToString(usernameHash[:]),
		}, nil
	}

	return nil, ErrUserNotFound
}
