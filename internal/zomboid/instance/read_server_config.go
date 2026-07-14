package instance

import (
	"context"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	serverconfig "github.com/zed-assistant/mcp/internal/zomboid/server_config"
)

type ReadServerConfigInput struct {
	InstanceID string `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance to read the config for" validate:"required"`
}

func (m *ZomboidInstanceManager) ReadServerConfig(ctx context.Context, principal authorization.Principal, input ReadServerConfigInput) (map[string]serverconfig.ConfigEntry, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	m.instanceLockManager.RLock(input.InstanceID)
	defer m.instanceLockManager.RUnlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	return m.serverConfigManager.ReadServerConfig(instanceCfg.HomeDir)
}
