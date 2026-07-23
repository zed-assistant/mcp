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

type ManagePlayerStateInput struct {
	InstanceID string `json:"instanceId" jsonschema:"ID of the Project Zomboid server instance to execute the command on" validate:"required"`
	Action     string `json:"action" jsonschema:"Action to perform on the player. Allowed actions: additem, addxp, invisible, noclip" validate:"required,oneof=additem addxp invisible noclip"`
	Username   string `json:"username" jsonschema:"Username of the player to manage" validate:"required"`
	ItemType   string `json:"itemType,omitempty" jsonschema:"Full item module ID to give the player (e.g. Base.Axe). Required for additem" validate:"omitempty"`
	Count      int    `json:"count,omitempty" jsonschema:"Number of items to add. Optional for additem, defaults to 1" validate:"omitempty,min=1"`
	Perk       string `json:"perk,omitempty" jsonschema:"Perk to grant XP in (e.g. Woodwork, Cooking, Axe). Not a fixed enum - the server validates it and returns the list of valid perks if unknown. Required for addxp" validate:"omitempty"`
	Amount     int    `json:"amount,omitempty" jsonschema:"Amount of XP to add, must be greater than 0. Required for addxp" validate:"omitempty"`
	Enabled    *bool  `json:"enabled,omitempty" jsonschema:"Whether to enable (true) or disable (false) the state. Required for invisible and noclip" validate:"omitempty"`
}

type ManagePlayerStateOutput struct {
	Action   string `json:"action" jsonschema:"The action performed on the player"`
	Username string `json:"username" jsonschema:"The username of the player the action was performed on"`
	Message  string `json:"message,omitempty" jsonschema:"Message returned by the server, if any"`
}

func (m *McpToolManager) ManagePlayerState() Tool {
	return &MCPTool[ManagePlayerStateInput, *ManagePlayerStateOutput]{
		Definition: &mcp.Tool{
			Name:        "manage-player-state",
			Description: "Manages a player's in-game state on the Project Zomboid server instance (additem, addxp, invisible, noclip).",
			Title:       "Manage Player State",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  false,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Manage Player State",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input ManagePlayerStateInput) (*ManagePlayerStateOutput, error) {
			var cmd admincommand.AdminCommand[string]

			switch input.Action {
			case "additem":
				if input.ItemType == "" {
					return nil, &domainerror.DomainError{
						InternalMessage: "additem action requires itemType",
						PublicMessage:   "additem action requires itemType",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewAddItemAdminCommand(input.Username, input.ItemType, input.Count)
			case "addxp":
				if input.Perk == "" {
					return nil, &domainerror.DomainError{
						InternalMessage: "addxp action requires perk",
						PublicMessage:   "addxp action requires perk",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				if input.Amount <= 0 {
					return nil, &domainerror.DomainError{
						InternalMessage: "addxp action requires amount greater than 0",
						PublicMessage:   "addxp action requires amount greater than 0",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewAddXPAdminCommand(input.Username, input.Perk, input.Amount)
			case "invisible":
				if input.Enabled == nil {
					return nil, &domainerror.DomainError{
						InternalMessage: "invisible action requires enabled",
						PublicMessage:   "invisible action requires enabled",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewInvisiblePlayerAdminCommand(input.Username, *input.Enabled)
			case "noclip":
				if input.Enabled == nil {
					return nil, &domainerror.DomainError{
						InternalMessage: "noclip action requires enabled",
						PublicMessage:   "noclip action requires enabled",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewNoclipAdminCommand(input.Username, *input.Enabled)
			default:
				return nil, &domainerror.DomainError{
					InternalMessage: fmt.Sprintf("Unsupported manage player state action %s", input.Action),
					PublicMessage:   fmt.Sprintf("Unsupported manage player state action %s. Allowed actions are: additem, addxp, invisible, noclip.", input.Action),
					InternalCode:    domainerror.InvalidInput,
				}
			}

			message, err := m.zomboidInstanceManager.ExecuteRawAdminCommand(ctx, principal, &instance.ExecuteAdminCommandInput[string]{
				InstanceID: input.InstanceID,
				Command:    cmd,
			})
			if err != nil {
				return nil, err
			}

			return &ManagePlayerStateOutput{
				Action:   input.Action,
				Username: input.Username,
				Message:  message,
			}, nil
		}),
	}
}
