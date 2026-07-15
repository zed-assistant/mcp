package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/config"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
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

type configType string

const (
	ConfigTypeServer configType = "server"
)

type ReadConfigInput struct {
	InstanceId string     `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance to read the configuration from." validate:"required"`
	ConfigType configType `json:"configType" jsonschema:"The type of configuration to read. Supported values: 'server' for server configuration." validate:"required,oneof=server"`
	Keys       *[]string  `json:"keys,omitempty" jsonschema:"Optional filter. Omit to return all. These files are large - filter when you know what you want. You can use * placeholder in any key filter."`
}

func (m *McpToolManager) ReadZomboidServerConfig() Tool {
	return &MCPTool[ReadConfigInput, map[string]config.ConfigEntry]{
		Definition: &mcp.Tool{
			Name:        "read-zomboid-config",
			Description: "Reads Project Zomboid configuration. Server config (ports, player limits, PVP, mods list)",
			Title:       "Read Project Zomboid config",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    true,
				Title:           "Read Project Zomboid config",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input ReadConfigInput) (map[string]config.ConfigEntry, error) {
			var keyFilters []string = nil
			if input.Keys != nil {
				keyFilters = *input.Keys
			}

			return m.zomboidInstanceManager.ReadServerConfig(ctx, principal, instance.ReadServerConfigInput{
				InstanceID: input.InstanceId,
				KeyFilters: keyFilters,
			})
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
