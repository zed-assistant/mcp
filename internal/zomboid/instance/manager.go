package instance

import (
	"context"
	"log/slog"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
	"github.com/zed-assistant/mcp/internal/zomboid/config"
	"github.com/zed-assistant/mcp/internal/zomboid/whitelist"
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
	ReadServerConfig(instanceConfig *configuration.ZomboidInstanceConfig, keysFilter []string) (map[string]config.ConfigEntry, error)
	UpdateServerConfig(instanceConfig *configuration.ZomboidInstanceConfig, newConfig map[string]string) error
}

type WhitelistManager interface {
	GetAllUsers(ctx context.Context, instanceConfig configuration.ZomboidInstanceConfig) ([]*whitelist.User, error)
}

type ZomboidInstanceManager struct {
	appConfig           *configuration.AppConfig
	instanceAuth        InstanceAuth
	instanceLockManager LockManager
	serverConfigManager ConfigManager
	log                 *slog.Logger
	adminCommandManager admincommand.AdminCommandManager
	whitelistManager    WhitelistManager
}

func NewZomboidInstanceManager(
	appConfig *configuration.AppConfig,
	instanceAuth InstanceAuth,
	instanceLockManager LockManager,
	serverConfigManager ConfigManager,
	log *slog.Logger,
	adminCommandManager admincommand.AdminCommandManager,
	whitelistManager WhitelistManager,
) *ZomboidInstanceManager {
	return &ZomboidInstanceManager{
		appConfig:           appConfig,
		instanceAuth:        instanceAuth,
		instanceLockManager: instanceLockManager,
		serverConfigManager: serverConfigManager,
		log:                 log,
		adminCommandManager: adminCommandManager,
		whitelistManager:    whitelistManager,
	}
}
