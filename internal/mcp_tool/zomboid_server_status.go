package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

type GetServerStatusInput struct {
	InstanceID string `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance" validate:"required"`
}

func (m *McpToolManager) GetServerStatus() Tool {
	return &MCPTool[GetServerStatusInput, *instance.ServerStatus]{
		Definition: &mcp.Tool{
			Name:        "get-server-status",
			Description: "Gets status of the Project Zomboid server instance: uptime, connected players with access levels. Cheap and safe - call this before any action that targets a player, to confirm they're online and to get their exact username.",
			Title:       "Get Project Zomboid Server Status",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    true,
				Title:           "Get Project Zomboid Server Status",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input GetServerStatusInput) (*instance.ServerStatus, error) {
			return m.zomboidInstanceManager.GetServerStatus(ctx, principal, &instance.GetServerStatusInput{
				InstanceID: input.InstanceID,
			})
		}),
	}
}
