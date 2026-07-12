package authapi

import (
	"fmt"

	"github.com/ory/fosite"
)

func oauthErrorToLoggerError(err error) error {
	rfcErr := fosite.ErrorToRFC6749Error(err)
	return fmt.Errorf("OAuth2 error: %s, description: %s, hint: %s, debug: %s", rfcErr.ErrorField, rfcErr.DescriptionField, rfcErr.HintField, rfcErr.DebugField)
}
