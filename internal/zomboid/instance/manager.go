package instance

import (
	"log/slog"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
	serverconfig "github.com/zed-assistant/mcp/internal/zomboid/server_config"
)

type InstanceAuth interface {
	AuthorizeInstanceAccess(instanceID string, principal authorization.Principal) error
}

type LockManager interface {
	Lock(instanceId string)
	Unlock(instanceId string)
	RLock(instanceId string)
	RUnlock(instanceId string)
}

type ServerConfigManager interface {
	ReadServerConfig(instanceID string) (map[string]serverconfig.ConfigEntry, error)
	UpdateConfig(instanceHomeDir string, newConfig map[string]string) error
}

type ZomboidInstanceManager struct {
	appConfig           *configuration.AppConfig
	instanceAuth        InstanceAuth
	instanceLockManager LockManager
	serverConfigManager ServerConfigManager
	log                 *slog.Logger
}

func NewZomboidInstanceManager(
	appConfig *configuration.AppConfig,
	instanceAuth InstanceAuth,
	instanceLockManager LockManager,
	serverConfigManager ServerConfigManager,
	log *slog.Logger,
) *ZomboidInstanceManager {
	return &ZomboidInstanceManager{
		appConfig:           appConfig,
		instanceAuth:        instanceAuth,
		instanceLockManager: instanceLockManager,
		serverConfigManager: serverConfigManager,
		log:                 log,
	}
}
