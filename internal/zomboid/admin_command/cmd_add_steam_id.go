package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type AddSteamIDAdminCommand struct {
	SteamID string
}

func (c AddSteamIDAdminCommand) ToCommand() string {
	return "addsteamid " + c.SteamID
}

func (c AddSteamIDAdminCommand) ParseResponse(response string) (string, error) {
	switch response {
	case fmt.Sprintf("SteamID %s added to allowed SteamIDs", c.SteamID):
		return "", nil
	case fmt.Sprintf("SteamID %s already exists in allowed SteamIDs", c.SteamID):
		return "", &domainerror.DomainError{
			InternalMessage: response,
			PublicMessage:   response,
			InternalCode:    domainerror.Conflict,
		}
	default:
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewAddSteamIDAdminCommand(steamID string) *AddSteamIDAdminCommand {
	return &AddSteamIDAdminCommand{
		SteamID: steamID,
	}
}
