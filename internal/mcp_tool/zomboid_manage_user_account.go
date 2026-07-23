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

type ManageUserAccountInput struct {
	InstanceID  string `json:"instanceId" jsonschema:"ID of the Project Zomboid server instance to execute the command on" validate:"required"`
	Action      string `json:"action" jsonschema:"Action to perform on the user account. Allowed actions: addsteamid, removesteamid, adduser, setaccesslevel, setpassword" validate:"required,oneof=addsteamid removesteamid adduser setaccesslevel setpassword"`
	SteamID     string `json:"steamId,omitempty" jsonschema:"Steam ID to add to or remove from the allowed SteamIDs list. Required for addsteamid and removesteamid" validate:"omitempty"`
	Username    string `json:"username,omitempty" jsonschema:"Username of the account to manage. Required for adduser, setaccesslevel and setpassword" validate:"omitempty"`
	Password    string `json:"password,omitempty" jsonschema:"Password for the account. Required for adduser and setpassword" validate:"omitempty"`
	AccessLevel string `json:"accessLevel,omitempty" jsonschema:"Access level to assign to the user. Required for setaccesslevel" validate:"omitempty"`
}

type ManageUserAccountOutput struct {
	Action  string `json:"action" jsonschema:"The action performed on the user account"`
	Message string `json:"message,omitempty" jsonschema:"Message returned by the server, if any (e.g. the generated password hash for setpassword)"`
}

func (m *McpToolManager) ManageUserAccount() Tool {
	return &MCPTool[ManageUserAccountInput, *ManageUserAccountOutput]{
		Definition: &mcp.Tool{
			Name:        "manage-user-account",
			Description: "Manages user accounts on the Project Zomboid server instance (addsteamid, removesteamid, adduser, setaccesslevel, setpassword).",
			Title:       "Manage User Account",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(true),
				IdempotentHint:  false,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    false,
				Title:           "Manage User Account",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input ManageUserAccountInput) (*ManageUserAccountOutput, error) {
			var cmd admincommand.AdminCommand[string]

			switch input.Action {
			case "addsteamid":
				if input.SteamID == "" {
					return nil, &domainerror.DomainError{
						InternalMessage: "addsteamid action requires steamId",
						PublicMessage:   "addsteamid action requires steamId",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewAddSteamIDAdminCommand(input.SteamID)
			case "removesteamid":
				if input.SteamID == "" {
					return nil, &domainerror.DomainError{
						InternalMessage: "removesteamid action requires steamId",
						PublicMessage:   "removesteamid action requires steamId",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewRemoveSteamIDAdminCommand(input.SteamID)
			case "adduser":
				if input.Username == "" || input.Password == "" {
					return nil, &domainerror.DomainError{
						InternalMessage: "adduser action requires username and password",
						PublicMessage:   "adduser action requires username and password",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewAddUserAdminCommand(input.Username, input.Password)
			case "setaccesslevel":
				if input.Username == "" || input.AccessLevel == "" {
					return nil, &domainerror.DomainError{
						InternalMessage: "setaccesslevel action requires username and accessLevel",
						PublicMessage:   "setaccesslevel action requires username and accessLevel",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewSetAccessLevelAdminCommand(input.Username, input.AccessLevel)
			case "setpassword":
				if input.Username == "" || input.Password == "" {
					return nil, &domainerror.DomainError{
						InternalMessage: "setpassword action requires username and password",
						PublicMessage:   "setpassword action requires username and password",
						InternalCode:    domainerror.InvalidInput,
					}
				}
				cmd = admincommand.NewSetPasswordAdminCommand(input.Username, input.Password)
			default:
				return nil, &domainerror.DomainError{
					InternalMessage: fmt.Sprintf("Unsupported manage user account action %s", input.Action),
					PublicMessage:   fmt.Sprintf("Unsupported manage user account action %s. Allowed actions are: addsteamid, removesteamid, adduser, setaccesslevel, setpassword.", input.Action),
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

			return &ManageUserAccountOutput{
				Action:  input.Action,
				Message: message,
			}, nil
		}),
	}
}
