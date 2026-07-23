package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type RemoveSteamIDAdminCommand struct {
	SteamID string
}

func (c RemoveSteamIDAdminCommand) ToCommand() string {
	return fmt.Sprintf("removesteamid %s", c.SteamID)
}

func (c RemoveSteamIDAdminCommand) ParseResponse(response string) (string, error) {
	if response == fmt.Sprintf("SteamID %s removed from allowed SteamIDs", c.SteamID) {
		return "", nil
	} else if response == fmt.Sprintf("SteamID %s doesn't exists in allowed SteamIDs", c.SteamID) {
		return "", &domainerror.DomainError{
			InternalMessage: response,
			PublicMessage:   response,
			InternalCode:    domainerror.NotFound,
		}
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewRemoveSteamIDAdminCommand(steamID string) *RemoveSteamIDAdminCommand {
	return &RemoveSteamIDAdminCommand{
		SteamID: steamID,
	}
}
