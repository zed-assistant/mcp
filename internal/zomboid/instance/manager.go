package instance

import (
	"log/slog"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/zomboid/config"
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

type ConfigManager interface {
	ReadServerConfig(instanceID string, keysFilter []string) (map[string]config.ConfigEntry, error)
	UpdateServerConfig(instanceHomeDir string, newConfig map[string]string) error
}

type ZomboidInstanceManager struct {
	appConfig           *configuration.AppConfig
	instanceAuth        InstanceAuth
	instanceLockManager LockManager
	serverConfigManager ConfigManager
	log                 *slog.Logger
}

func NewZomboidInstanceManager(
	appConfig *configuration.AppConfig,
	instanceAuth InstanceAuth,
	instanceLockManager LockManager,
	serverConfigManager ConfigManager,
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
