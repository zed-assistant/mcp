package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

func (m *McpToolManger) ListZomboidInstances() Tool {
	return &MCPTool[EmptyInput, []*instance.Instance]{
		Definition: &mcp.Tool{
			Name:        "list-zomboid-instances",
			Description: "Lists Project Zomboid server instances available for user",
		},
		Handler: withUserRecoverNoInput(m.logger, func(ctx context.Context, principal authorization.Principal) ([]*instance.Instance, error) {
			return m.zomboidInstanceManager.ListInstances(ctx, principal)
		}),
	}
}
