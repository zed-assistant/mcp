package configuration

import (
	"errors"
	"fmt"
)

func (c *AppConfig) validate() error {
	var errs []error

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be between 1 and 65535"))
	}

	if len(c.OAuth2.SigningSecret) > 0 && len(c.OAuth2.SigningSecret) != 32 {
		errs = append(errs, fmt.Errorf("oauth2.signing_secret must be 32 characters long"))
	}

	if c.Server.ExternalUrl == "" {
		errs = append(errs, fmt.Errorf("server.external_url must be set"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
