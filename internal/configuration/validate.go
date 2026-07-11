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

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
