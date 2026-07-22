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

type ModeratePlayerInput struct {
	InstanceID string `json:"instanceId" jsonschema:"ID of the Project Zomboid server instance to execute the command on" validate:"required"`
	Action     string `json:"action" jsonschema:"Action to perform on the player. Allowed actions: kick, ban, unban" validate:"required,oneof=kick ban unban"`
	TargetType string `json:"targetType" jsonschema:"Type of the target player. Allowed types: steamId, username" validate:"required,oneof=steamId username"`
	Target     string `json:"target" jsonschema:"The target player's identifier (Steam ID or username) based on the specified targetType" validate:"required"`
	Reason     string `json:"reason,omitempty" jsonschema:"Reason for the action (optional, but strongly recommended)" validate:"omitempty"`
}

type ModeratePlayerOutput struct {
	Action     string                 `json:"action" jsonschema:"The action performed on the player"`
	TargetType string                 `json:"targetType" jsonschema:"The type of the target player (steamId or username)"`
	Target     string                 `json:"target" jsonschema:"The target player's identifier (Steam ID or username)"`
	Status     *instance.ServerStatus `json:"status" jsonschema:"The current status of the server after the action"`
}

func (m *McpToolManager) ModeratePlayer() Tool {
	return &MCPTool[ModeratePlayerInput, *ModeratePlayerOutput]{
		Definition: &mcp.Tool{
			Name:        "moderate-player",
			Description: "Moderates a player on the Project Zomboid server instance (kick, ban, unban).",
			Title:       "Moderate Player",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Moderate Player",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input ModeratePlayerInput) (*ModeratePlayerOutput, error) {
			var err error

			switch input.Action {
			case "kick":
				if input.TargetType != "username" {
					return nil, &domainerror.DomainError{
						InternalMessage: fmt.Sprintf("Kick action requires targetType to be 'username', got '%s'", input.TargetType),
						PublicMessage:   fmt.Sprintf("Kick action requires targetType to be 'username', got '%s'", input.TargetType),
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd := admincommand.NewKickUserAdminCommand(input.Target, input.Reason)
				_, err = m.zomboidInstanceManager.ExecuteRawAdminCommand(ctx, principal, &instance.ExecuteAdminCommandInput[string]{
					InstanceID: input.InstanceID,
					Command:    cmd,
				})
			case "ban":
				var cmd admincommand.AdminCommand[string]
				if input.TargetType == "username" {
					cmd = admincommand.NewBanUserAdminCommand(input.Target, input.Reason)
				} else {
					cmd = admincommand.NewBanUserIDAdminCommand(input.Target)
				}
				_, err = m.zomboidInstanceManager.ExecuteRawAdminCommand(ctx, principal, &instance.ExecuteAdminCommandInput[string]{
					InstanceID: input.InstanceID,
					Command:    cmd,
				})
			case "unban":
				var cmd admincommand.AdminCommand[string]
				if input.TargetType == "username" {
					cmd = admincommand.NewUnbanUserAdminCommand(input.Target)
				} else {
					cmd = admincommand.NewUnbanUserIDAdminCommand(input.Target)
				}
				_, err = m.zomboidInstanceManager.ExecuteRawAdminCommand(ctx, principal, &instance.ExecuteAdminCommandInput[string]{
					InstanceID: input.InstanceID,
					Command:    cmd,
				})
			default:
				return nil, &domainerror.DomainError{
					InternalMessage: fmt.Sprintf("Unsupported moderate user action %s", input.Action),
					PublicMessage:   fmt.Sprintf("Unsupported moderate user action %s. Allowed actions are: kick, ban, unban.", input.Action),
					InternalCode:    domainerror.InvalidInput,
				}
			}

			serverStatus, err := m.zomboidInstanceManager.GetServerStatus(ctx, principal, &instance.GetServerStatusInput{InstanceID: input.InstanceID})
			if err != nil {
				return nil, fmt.Errorf("failed to get server status after moderating player: %w", err)
			}

			return &ModeratePlayerOutput{
				Action:     input.Action,
				TargetType: input.TargetType,
				Target:     input.Target,
				Status:     serverStatus,
			}, err
		}),
	}
}
