package instance

import (
	"context"
	"fmt"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
	"github.com/zed-assistant/mcp/internal/logger"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
	"github.com/zed-assistant/mcp/internal/zomboid/config"
)

type ReadConfigInput struct {
	InstanceID string
	ConfigType config.ConfigType
	KeyFilters []string
}

func (m *ZomboidInstanceManager) ReadConfig(ctx context.Context, principal authorization.Principal, input ReadConfigInput) (map[string]any, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	m.instanceLockManager.RLock(input.InstanceID)
	defer m.instanceLockManager.RUnlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	switch input.ConfigType {
	case config.ConfigTypeSandbox:
		return m.serverConfigManager.ReadSandboxConfig(&instanceCfg, input.KeyFilters)
	case config.ConfigTypeServer:
		entries, err := m.serverConfigManager.ReadServerConfig(&instanceCfg, input.KeyFilters)
		if err != nil {
			return nil, err
		}
		return serverEntriesToAnyMap(entries), nil
	default:
		return nil, unsupportedConfigTypeError(input.ConfigType)
	}
}

type UpdateConfigInput struct {
	InstanceID string
	ConfigType config.ConfigType
	Updates    map[string]any
	ApplyLive  bool
}

func (m *ZomboidInstanceManager) UpdateConfig(ctx context.Context, principal authorization.Principal, input UpdateConfigInput) (map[string]any, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	m.instanceLockManager.Lock(input.InstanceID)
	defer m.instanceLockManager.Unlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	switch input.ConfigType {
	case config.ConfigTypeServer:
		return m.updateServerConfig(ctx, principal, instanceCfg, input)
	case config.ConfigTypeSandbox:
		return m.updateSandboxConfig(ctx, principal, instanceCfg, input)
	default:
		return nil, unsupportedConfigTypeError(input.ConfigType)
	}
}

func (m *ZomboidInstanceManager) updateServerConfig(ctx context.Context, _ authorization.Principal, instanceCfg configuration.ZomboidInstanceConfig, input UpdateConfigInput) (map[string]any, error) {
	flatUpdates, err := config.FlattenServerConfigUpdates(input.Updates)
	if err != nil {
		return nil, err
	}

	m.log.InfoContext(ctx, fmt.Sprintf("Updating server config for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	if err := m.serverConfigManager.UpdateServerConfig(&instanceCfg, flatUpdates); err != nil {
		m.log.ErrorContext(ctx, "Server config update failed", logger.LogError(err))
		return nil, err
	}

	m.log.InfoContext(ctx, fmt.Sprintf("Server config updated successfully for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	if input.ApplyLive {
		m.log.InfoContext(ctx, fmt.Sprintf("Applying server config changes live for instance %s (%s)", input.InstanceID, instanceCfg.Name))
		_, err := admincommand.ExecuteSingleAdminCommand(m.adminCommandManager, &instanceCfg, &admincommand.ReloadOptionsAdminCommand{})
		if err != nil {
			m.log.ErrorContext(ctx, "Failed to apply server config changes live", logger.LogError(err))
			return nil, err
		}
		m.log.InfoContext(ctx, fmt.Sprintf("Server config changes applied live successfully for instance %s (%s)", input.InstanceID, instanceCfg.Name))
	}

	entries, err := m.serverConfigManager.ReadServerConfig(&instanceCfg, nil)
	if err != nil {
		return nil, err
	}
	return serverEntriesToAnyMap(entries), nil
}

func (m *ZomboidInstanceManager) updateSandboxConfig(ctx context.Context, _ authorization.Principal, instanceCfg configuration.ZomboidInstanceConfig, input UpdateConfigInput) (map[string]any, error) {
	if input.ApplyLive {
		return nil, &domainerror.DomainError{
			InternalMessage: "applyLive is not supported for configType=sandbox",
			PublicMessage:   "applyLive is not supported for configType=sandbox: there is no known way to confirm sandbox variable changes apply to a running server, and many of them only take effect on world/server restart. Save the change (applyLive=false) and restart when appropriate.",
			InternalCode:    domainerror.InvalidInput,
		}
	}

	m.log.InfoContext(ctx, fmt.Sprintf("Updating sandbox config for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	if err := m.serverConfigManager.UpdateSandboxConfig(&instanceCfg, input.Updates); err != nil {
		m.log.ErrorContext(ctx, "Sandbox config update failed", logger.LogError(err))
		return nil, err
	}

	m.log.InfoContext(ctx, fmt.Sprintf("Sandbox config updated successfully for instance %s (%s)", input.InstanceID, instanceCfg.Name))

	return m.serverConfigManager.ReadSandboxConfig(&instanceCfg, nil)
}

func serverEntriesToAnyMap(entries map[string]config.ConfigEntry) map[string]any {
	result := make(map[string]any, len(entries))
	for key, entry := range entries {
		result[key] = entry
	}
	return result
}

func unsupportedConfigTypeError(configType config.ConfigType) error {
	return &domainerror.DomainError{
		InternalMessage: fmt.Sprintf("unsupported config type: %s", configType),
		PublicMessage:   fmt.Sprintf("unsupported config type: %s", configType),
		InternalCode:    domainerror.InvalidInput,
	}
}
