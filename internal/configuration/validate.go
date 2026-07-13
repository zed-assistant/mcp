package configuration

import (
	"errors"
	"fmt"
	"net/mail"
	"slices"
)

var allowedIDPTypes = []string{"local"}

func (c *AppConfig) validate() error {
	var errs []error

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be between 1 and 65535"))
	}

	if c.Server.ExternalUrl == "" {
		errs = append(errs, fmt.Errorf("server.external_url must be set"))
	}

	if len(c.OAuth2.SigningSecret) > 0 && len(c.OAuth2.SigningSecret) != 32 {
		errs = append(errs, fmt.Errorf("oauth2.signing_secret must be 32 characters long"))
	}

	if !slices.Contains(allowedIDPTypes, c.OAuth2.IDP.Type) {
		errs = append(errs, fmt.Errorf("oauth2.idp.type %s is not allowed - must be one of: %v", c.OAuth2.IDP.Type, allowedIDPTypes))
	}

	switch c.OAuth2.IDP.Type {
	case "local":
		errs = append(errs, c.validateLocalIDP()...)
	}

	if len(c.Zomboid.Instances) == 0 {
		errs = append(errs, fmt.Errorf("zomboid.instances must contain at least one instance"))
	}

	for id, instance := range c.Zomboid.Instances {
		if id == "" {
			errs = append(errs, fmt.Errorf("zomboid.instances[%s].id must be set", id))
		}
		if instance.Name == "" {
			errs = append(errs, fmt.Errorf("zomboid.instances[%s].name must be set", id))
		}
		if instance.HomeDir == "" {
			errs = append(errs, fmt.Errorf("zomboid.instances[%s].home_dir must be set", id))
		}
		if len(instance.Users) == 0 {
			errs = append(errs, fmt.Errorf("zomboid.instances[%s].users must contain at least one user", id))
		}

		for j, user := range instance.Users {
			if user == "" {
				errs = append(errs, fmt.Errorf("zomboid.instances[%s].users[%d] must be set", id, j))
			}

			if user != "" && !isValidEmail(user) {
				errs = append(errs, fmt.Errorf("zomboid.instances[%s].users[%d] must be a valid email address", id, j))
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (c *AppConfig) validateLocalIDP() []error {
	var errs []error

	cfg := c.OAuth2.IDP.Local

	if cfg == nil {
		errs = append(errs, fmt.Errorf("oauth2.idp.local must be set when oauth2.idp.type is 'local'"))
	}
	if len(cfg.Users) == 0 {
		errs = append(errs, fmt.Errorf("oauth2.idp.local.users must contain at least one user when oauth2.idp.type is 'local'"))
	}

	for i, user := range cfg.Users {
		if user.Username == "" {
			errs = append(errs, fmt.Errorf("oauth2.idp.local.users[%d].username must be set", i))
		}
		if user.Username != "" && !isValidEmail(user.Username) {
			errs = append(errs, fmt.Errorf("oauth2.idp.local.users[%d].username must be a valid email address", i))
		}
		if user.Password == "" {
			errs = append(errs, fmt.Errorf("oauth2.idp.local.users[%d].password must be set", i))
		}
	}

	return errs
}

func isValidEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s
}
