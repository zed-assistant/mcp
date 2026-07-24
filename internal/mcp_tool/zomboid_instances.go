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

type ReadConfigInput struct {
	InstanceId string            `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance to read the configuration from." validate:"required"`
	ConfigType config.ConfigType `json:"configType" jsonschema:"The type of configuration to read. 'server' is the flat ServerOptions .ini (ports, player limits, PVP, mods list). 'sandbox' is the SandboxVars.lua world/gameplay settings (zombies, loot, time, etc.), which may be arbitrarily nested and can include entries added by mods." validate:"required,oneof=server sandbox"`
	Keys       *[]string         `json:"keys,omitempty" jsonschema:"Optional filter. Omit to return all. These files are large - filter when you know what you want. Filters are matched against the full dotted path of each setting (e.g. 'ZombieLore.Speed'), and you can use * as a wildcard anywhere in a filter (e.g. 'ZombieLore.*' to select a whole group, or '*Speed*' to find similarly named settings anywhere)."`
}

func (m *McpToolManager) ReadZomboidConfig() Tool {
	return &MCPTool[ReadConfigInput, map[string]any]{
		Definition: &mcp.Tool{
			Name: "read-zomboid-config",
			Description: "Reads Project Zomboid configuration. For configType='server' this returns a flat object of " +
				"{ key: { value, description } }. For configType='sandbox' this returns a nested object mirroring the " +
				"SandboxVars table structure: leaf settings are { value, description, type }, and groups (e.g. 'ZombieLore', " +
				"'MultiplierConfig') are plain nested objects containing further settings - note the same short name can " +
				"exist in more than one group with a different meaning (e.g. top-level 'Farming' vs 'MultiplierConfig.Farming'), " +
				"so always read the full nesting, not just the leaf name. 'type' ('integer'/'number'/'boolean'/'string') is only " +
				"present for configType='sandbox' and tells you what an update's value must look like for that setting - note " +
				"'integer' (whole number, no decimal point) and 'number' (float, decimal point allowed) are distinct: an " +
				"integer setting rejects a fractional update value. 'type' is absent for configType='server', where every " +
				"value is always a plain string.",
			Title: "Read Project Zomboid config",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    true,
				Title:           "Read Project Zomboid config",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input ReadConfigInput) (map[string]any, error) {
			var keyFilters []string = nil
			if input.Keys != nil {
				keyFilters = *input.Keys
			}

			return m.zomboidInstanceManager.ReadConfig(ctx, principal, instance.ReadConfigInput{
				InstanceID: input.InstanceId,
				ConfigType: input.ConfigType,
				KeyFilters: keyFilters,
			})
		}),
	}
}

type UpdateConfigInput struct {
	InstanceId string            `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance to update the configuration for." validate:"required"`
	ConfigType config.ConfigType `json:"configType" jsonschema:"The type of configuration to update. 'server' for the flat ServerOptions .ini, 'sandbox' for the SandboxVars.lua world/gameplay settings." validate:"required,oneof=server sandbox"`
	Updates    map[string]any    `json:"updates" jsonschema:"Partial update, matching the shape returned by read-zomboid-config for the same configType. For configType='server' this is a flat object of key -> new value string, e.g. {\"MaxPlayers\": \"32\"}. For configType='sandbox' this can be nested to reach settings inside groups, e.g. {\"Zombies\": \"4\", \"ZombieLore\": {\"Speed\": \"2\"}}. Every value must be a plain JSON string (even for numeric or boolean settings - e.g. \"true\", \"4.5\"); the existing setting's type is preserved automatically. If ANY key is unknown, or ANY value has the wrong shape/type for its setting, the ENTIRE update is rejected with an error listing every problem found, and NOTHING is changed."`
	ApplyLive  bool              `json:"applyLive,omitempty" jsonschema:"Attempt to apply the updates live to the running server via RCON reloadoptions. Only supported for configType='server' (works for the .ini ServerOptions, e.g. ports, player limits). Not supported for configType='sandbox' and will be rejected: there is no confirmed way to hot-reload SandboxVars, and many sandbox settings only take effect on world/server restart. If false or missing, updates are saved and applied on next server restart."`
}

func (m *McpToolManager) UpdateZomboidConfig() Tool {
	return &MCPTool[UpdateConfigInput, map[string]any]{
		Definition: &mcp.Tool{
			Name: "update-zomboid-config",
			Description: "Updates Project Zomboid configuration for a given instance (server .ini or sandbox SandboxVars.lua, " +
				"selected via configType). The input is a partial update: only the provided keys are changed, everything else " +
				"(including comments/descriptions in the sandbox lua file) is left untouched. The whole update is validated " +
				"before anything is written - if any key is unknown or any value doesn't match its setting's expected shape " +
				"or type, no changes are applied at all and a combined error listing every problem is returned.",
			Title: "Update Project Zomboid config",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(true),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Update Project Zomboid config",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input UpdateConfigInput) (map[string]any, error) {
			return m.zomboidInstanceManager.UpdateConfig(ctx, principal, instance.UpdateConfigInput{
				InstanceID: input.InstanceId,
				ConfigType: input.ConfigType,
				Updates:    input.Updates,
				ApplyLive:  input.ApplyLive,
			})
		}),
	}
}
