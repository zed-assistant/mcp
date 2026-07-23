package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type AddSteamIDAdminCommand struct {
	SteamID string
}

func (c AddSteamIDAdminCommand) ToCommand() string {
	return fmt.Sprintf("addsteamid %s", c.SteamID)
}

func (c AddSteamIDAdminCommand) ParseResponse(response string) (string, error) {
	if response == fmt.Sprintf("SteamID %s added to allowed SteamIDs", c.SteamID) {
		return "", nil
	} else if response == fmt.Sprintf("SteamID %s already exists in allowed SteamIDs", c.SteamID) {
		return "", &domainerror.DomainError{
			InternalMessage: response,
			PublicMessage:   response,
			InternalCode:    domainerror.Conflict,
		}
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewAddSteamIDAdminCommand(steamID string) *AddSteamIDAdminCommand {
	return &AddSteamIDAdminCommand{
		SteamID: steamID,
	}
}
