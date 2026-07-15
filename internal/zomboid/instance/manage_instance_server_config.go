package instance

import (
	"context"
	"fmt"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/logger"
	"github.com/zed-assistant/mcp/internal/zomboid/config"
)

type ReadServerConfigInput struct {
	InstanceID string
	KeyFilters []string
}

func (m *ZomboidInstanceManager) ReadServerConfig(ctx context.Context, principal authorization.Principal, input ReadServerConfigInput) (map[string]config.ConfigEntry, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	m.instanceLockManager.RLock(input.InstanceID)
	defer m.instanceLockManager.RUnlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	return m.serverConfigManager.ReadServerConfig(instanceCfg.HomeDir, input.KeyFilters)
}

type UpdateServerConfigInput struct {
	InstanceID string
	Updates    map[string]string
	ApplyLive  bool
}

func (m *ZomboidInstanceManager) UpdateServerConfig(ctx context.Context, principal authorization.Principal, input UpdateServerConfigInput) (map[string]config.ConfigEntry, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	m.instanceLockManager.Lock(input.InstanceID)
	defer m.instanceLockManager.Unlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	m.log.InfoContext(ctx, fmt.Sprintf("Updating server config for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	if err := m.serverConfigManager.UpdateServerConfig(instanceCfg.HomeDir, input.Updates); err != nil {
		m.log.Error("Server config update failed", logger.LogError(err))
		return nil, err
	}

	m.log.InfoContext(ctx, fmt.Sprintf("Server config updated successfully for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	return m.serverConfigManager.ReadServerConfig(instanceCfg.HomeDir, nil)
}
