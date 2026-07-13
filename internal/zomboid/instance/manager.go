package instance

import (
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
)

type InstanceAuth interface {
	AuthorizeInstanceAccess(instanceID string, principal authorization.Principal) error
}
type ZomboidInstanceManager struct {
	appConfig    *configuration.AppConfig
	instanceAuth InstanceAuth
}

func NewZomboidInstanceManager(appConfig *configuration.AppConfig, instanceAuth InstanceAuth) *ZomboidInstanceManager {
	return &ZomboidInstanceManager{
		appConfig:    appConfig,
		instanceAuth: instanceAuth,
	}
}
