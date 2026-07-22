package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

type BroadcastServerMessageInput struct {
	InstanceID string `json:"instanceId" jsonschema:"ID of the Project Zomboid server instance to execute the command on" validate:"required"`
	Message    string `json:"message" jsonschema:"Message to broadcast to all players on the server" validate:"required"`
}

func (m *McpToolManager) BroadcastServerMessage() Tool {
	return &MCPTool[BroadcastServerMessageInput, Empty]{
		Definition: &mcp.Tool{
			Name:        "broadcast-server-message",
			Description: "Broadcasts a message to all players on the Project Zomboid server instance.",
			Title:       "Broadcast Server Message",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  false,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Broadcast Server Message",
			},
		},
		Handler: withUserRecoverNoOutput(m.logger, func(ctx context.Context, principal authorization.Principal, input BroadcastServerMessageInput) error {
			cmd := &admincommand.ServerMessageAdminCommand{
				Message: input.Message,
			}

			_, err := m.zomboidInstanceManager.ExecuteRawAdminCommand(ctx, principal, &instance.ExecuteAdminCommandInput[string]{
				InstanceID: input.InstanceID,
				Command:    cmd,
			})
			return err
		}),
	}
}
