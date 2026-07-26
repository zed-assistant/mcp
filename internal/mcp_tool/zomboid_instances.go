package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

func (m *McpToolManager) ListZomboidInstances() *MCPTool[Empty, []*instance.Instance] {
	return &MCPTool[Empty, []*instance.Instance]{
		Definition: &mcp.Tool{
			Name:        "list-zomboid-instances",
			Description: "Lists Project Zomboid server instances available for user",
			Title:       "List Project Zomboid server instances",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    true,
				Title:           "List Project Zomboid server instances",
			},
		},
		Handler: withUserRecoverNoInput(m.logger, func(ctx context.Context, principal authorization.Principal) ([]*instance.Instance, error) {
			return m.zomboidInstanceManager.ListInstances(ctx, principal)
		}),
	}
}
