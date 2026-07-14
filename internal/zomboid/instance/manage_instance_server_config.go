package instance

import (
	"context"
	"fmt"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/logger"
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

type UpdateServerConfigInput struct {
	InstanceID string            `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance to update the config for" validate:"required"`
	Updates    map[string]string `json:"updates" jsonschema:"The partial updates as key-value pairs to apply to the server config"`
}

func (m *ZomboidInstanceManager) UpdateServerConfig(ctx context.Context, principal authorization.Principal, input UpdateServerConfigInput) error {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return err
	}

	m.instanceLockManager.Lock(input.InstanceID)
	defer m.instanceLockManager.Unlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	m.log.InfoContext(ctx, fmt.Sprintf("Updating server config for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	if err := m.serverConfigManager.UpdateConfig(instanceCfg.HomeDir, input.Updates); err != nil {
		m.log.Error("Server config update failed", logger.LogError(err))
		return err
	}

	m.log.InfoContext(ctx, fmt.Sprintf("Server config updated successfully for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	return nil
}
