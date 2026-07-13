package instanceauth

import (
	"fmt"
	"slices"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type InstanceAuthorization struct {
	appConfig *configuration.AppConfig
}

func NewInstanceAuthorization(appConfig *configuration.AppConfig) *InstanceAuthorization {
	return &InstanceAuthorization{
		appConfig: appConfig,
	}
}

func forbiddenError(internalMessage string) *domainerror.DomainError {
	return &domainerror.DomainError{
		InternalMessage: internalMessage,
		PublicMessage:   "User have no access to the requested Project Zomboid server instance",
		InternalCode:    domainerror.NotAllowed,
	}
}

func (ia *InstanceAuthorization) AuthorizeInstanceAccess(instanceID string, principal authorization.Principal) error {
	instance, ok := ia.appConfig.Zomboid.Instances[instanceID]
	if !ok {
		return forbiddenError(fmt.Sprintf("Instance %s not found", instanceID))
	}

	if slices.Contains(instance.Users, principal.Email) {
		return nil
	}

	return forbiddenError(fmt.Sprintf("User %s has no access to instance %s", principal.Subject, instanceID))
}
