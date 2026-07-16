package instance

import (
	"context"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
)

type ExecuteAdminCommandInput[T any] struct {
	InstanceID string
	Command    admincommand.AdminCommand[T]
}

func (m *ZomboidInstanceManager) ExecuteRawAdminCommand(ctx context.Context, principal authorization.Principal, input *ExecuteAdminCommandInput[string]) (string, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return "", err
	}

	m.instanceLockManager.Lock(input.InstanceID)
	defer m.instanceLockManager.Unlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	return admincommand.ExecuteSingleAdminCommand(m.adminCommandManager, &instanceCfg, input.Command)
}
