package mcptool

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

type ManageServerLifecycleInput struct {
	InstanceID string `json:"instanceId" jsonschema:"ID of the Project Zomboid server instance to execute the command on" validate:"required"`
	Action     string `json:"action" jsonschema:"Action to perform on the server. Allowed actions: reloadalllua, save, quit" validate:"required,oneof=reloadalllua save quit"`
	Confirm    bool   `json:"confirm,omitempty" jsonschema:"Must be explicitly set to true to execute the quit action, since it immediately shuts down the server and disconnects all players. Not required for other actions" validate:"omitempty"`
}

type ManageServerLifecycleOutput struct {
	Action  string `json:"action" jsonschema:"The action performed on the server"`
	Message string `json:"message,omitempty" jsonschema:"Message returned by the server, if any"`
}

func (m *McpToolManager) ManageServerLifecycle() Tool {
	return &MCPTool[ManageServerLifecycleInput, *ManageServerLifecycleOutput]{
		Definition: &mcp.Tool{
			Name:        "manage-server-lifecycle",
			Description: "Manages the lifecycle of the Project Zomboid server instance (reloadalllua, save, quit). The quit action requires confirm to be explicitly set to true, since it shuts down the server for all players.",
			Title:       "Manage Server Lifecycle",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(true),
				IdempotentHint:  false,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Manage Server Lifecycle",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input ManageServerLifecycleInput) (*ManageServerLifecycleOutput, error) {
			var cmd admincommand.AdminCommand[string]

			switch input.Action {
			case "reloadalllua":
				cmd = admincommand.NewReloadAllLuaAdminCommand()
			case "save":
				cmd = admincommand.NewSaveWorldAdminCommand()
			case "quit":
				if !input.Confirm {
					return nil, &domainerror.DomainError{
						InternalMessage: "quit action requires confirm to be true",
						PublicMessage:   "quit action requires confirm to be set to true, since it will shut down the server and disconnect all players",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewQuitServerAdminCommand()
			default:
				return nil, &domainerror.DomainError{
					InternalMessage: fmt.Sprintf("Unsupported manage server lifecycle action %s", input.Action),
					PublicMessage:   fmt.Sprintf("Unsupported manage server lifecycle action %s. Allowed actions are: reloadalllua, save, quit.", input.Action),
					InternalCode:    domainerror.InvalidInput,
				}
			}

			message, err := m.zomboidInstanceManager.ExecuteRawAdminCommand(ctx, principal, &instance.ExecuteAdminCommandInput[string]{
				InstanceID: input.InstanceID,
				Command:    cmd,
			})
			if err != nil {
				return nil, err
			}

			return &ManageServerLifecycleOutput{
				Action:  input.Action,
				Message: message,
			}, nil
		}),
	}
}
