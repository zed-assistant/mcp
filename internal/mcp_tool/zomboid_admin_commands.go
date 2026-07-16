package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

type RawAdminCommandInput struct {
	InstanceID string `json:"instanceId" jsonschema:"ID of the Project Zomboid server instance to execute the command on" validate:"required"`
	Command    string `json:"command" jsonschema:"Full admin command with arguments in a single string" validate:"required"`
}

func (m *McpToolManager) ExecuteRawAdminCommand() Tool {
	return &MCPTool[RawAdminCommandInput, string]{
		Definition: &mcp.Tool{
			Name:        "execute-raw-admin-command",
			Description: "Executes a raw Project Zomboid admin command on the server instance. Escape hatch for commands not covered by other tools - prefer the dedicated tool when one exists, since raw commands are unvalidated. Use 'help' command to see available commands.",
			Title:       "Execute raw Project Zomboid admin command",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(true),
				IdempotentHint:  false,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Execute raw Project Zomboid admin command",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input RawAdminCommandInput) (string, error) {
			cmd := &admincommand.RawAdminCommand{
				Cmd: input.Command,
			}

			return m.zomboidInstanceManager.ExecuteRawAdminCommand(ctx, principal, &instance.ExecuteAdminCommandInput[string]{
				InstanceID: input.InstanceID,
				Command:    cmd,
			})
		}),
	}
}
