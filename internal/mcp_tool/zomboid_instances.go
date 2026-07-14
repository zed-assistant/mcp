package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
	serverconfig "github.com/zed-assistant/mcp/internal/zomboid/server_config"
)

func (m *McpToolManager) ListZomboidInstances() Tool {
	return &MCPTool[EmptyInput, []*instance.Instance]{
		Definition: &mcp.Tool{
			Name:        "list-zomboid-instances",
			Description: "Lists Project Zomboid server instances available for user",
			Title:       "List Project Zomboid server instances",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  false,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "List Project Zomboid server instances",
			},
		},
		Handler: withUserRecoverNoInput(m.logger, func(ctx context.Context, principal authorization.Principal) ([]*instance.Instance, error) {
			return m.zomboidInstanceManager.ListInstances(ctx, principal)
		}),
	}
}

func (m *McpToolManager) ReadZomboidServerConfig() Tool {
	return &MCPTool[instance.ReadServerConfigInput, map[string]serverconfig.ConfigEntry]{
		Definition: &mcp.Tool{
			Name:        "read-zomboid-server-config",
			Description: "Reads Project Zomboid server config for a given instance",
			Title:       "Read Project Zomboid server config",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  false,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Read Project Zomboid server config",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input instance.ReadServerConfigInput) (map[string]serverconfig.ConfigEntry, error) {
			return m.zomboidInstanceManager.ReadServerConfig(ctx, principal, input)
		}),
	}
}
