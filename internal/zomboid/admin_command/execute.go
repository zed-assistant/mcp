package admincommand

import "github.com/zed-assistant/mcp/internal/configuration"

type AdminCommandManager interface {
	ExecuteCommand(instanceConfig *configuration.ZomboidInstanceConfig, command string) (string, error)
}

func ExecuteSingleAdminCommand[T any](svc AdminCommandManager, instanceConfig *configuration.ZomboidInstanceConfig, cmd AdminCommand[T]) (T, error) {
	response, err := svc.ExecuteCommand(instanceConfig, cmd.ToCommand())
	if err != nil {
		var zero T
		return zero, err
	}
	return cmd.ParseResponse(response)
}
