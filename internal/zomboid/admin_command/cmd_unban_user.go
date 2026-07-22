package admincommand

import (
	"fmt"
	"strings"
)

type UnbanUserAdminCommand struct {
	Username string
}

func (c UnbanUserAdminCommand) ToCommand() string {
	return "unbanuser \"" + c.Username + "\""
}

func (c UnbanUserAdminCommand) ParseResponse(response string) (string, error) {
	if strings.HasSuffix(response, fmt.Sprintf("unbanned user %s", c.Username)) {
		return "", nil
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewUnbanUserAdminCommand(username string) *UnbanUserAdminCommand {
	return &UnbanUserAdminCommand{
		Username: username,
	}
}

type UnbanUserIDAdminCommand struct {
	SteamID string
}

func (c UnbanUserIDAdminCommand) ToCommand() string {
	return "unbanid \"" + c.SteamID + "\""
}

func (c UnbanUserIDAdminCommand) ParseResponse(response string) (string, error) {
	if strings.ToLower(response) == fmt.Sprintf("steamid %s is now unbanned", strings.ToLower(c.SteamID)) {
		return "", nil
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewUnbanUserIDAdminCommand(steamID string) *UnbanUserIDAdminCommand {
	return &UnbanUserIDAdminCommand{
		SteamID: steamID,
	}
}
