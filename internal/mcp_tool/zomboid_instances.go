package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
	serverconfig "github.com/zed-assistant/mcp/internal/zomboid/server_config"
)

func (m *McpToolManager) ListZomboidInstances() Tool {
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

func (m *McpToolManager) ReadZomboidServerConfig() Tool {
	return &MCPTool[instance.ReadServerConfigInput, map[string]serverconfig.ConfigEntry]{
		Definition: &mcp.Tool{
			Name:        "read-zomboid-server-config",
			Description: "Reads Project Zomboid server config for a given instance",
			Title:       "Read Project Zomboid server config",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    true,
				Title:           "Read Project Zomboid server config",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input instance.ReadServerConfigInput) (map[string]serverconfig.ConfigEntry, error) {
			return m.zomboidInstanceManager.ReadServerConfig(ctx, principal, input)
		}),
	}
}

func (m *McpToolManager) UpdateZomboidServerConfig() Tool {
	return &MCPTool[instance.UpdateServerConfigInput, Empty]{
		Definition: &mcp.Tool{
			Name:        "update-zomboid-server-config",
			Description: "Updates Project Zomboid server config for a given instance. Provided input is a partial update, meaning that only the provided keys will be updated, and the rest of the config will remain unchanged.",
			Title:       "Update Project Zomboid server config",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(true),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Update Project Zomboid server config",
			},
		},
		Handler: withUserRecoverNoOutput(m.logger, func(ctx context.Context, principal authorization.Principal, input instance.UpdateServerConfigInput) error {
			return m.zomboidInstanceManager.UpdateServerConfig(ctx, principal, input)
		}),
	}
}
