package admincommand

import (
	"fmt"
	"strings"
)

type BanUserAdminCommand struct {
	Username string
	Reason   string
}

func (c BanUserAdminCommand) ToCommand() string {
	command := fmt.Sprintf("banuser \"%s\"", c.Username)
	if c.Reason != "" {
		command += fmt.Sprintf(" -r \"%s\"", c.Reason)
	}
	return command
}

func (c BanUserAdminCommand) ParseResponse(response string) (string, error) {
	if strings.HasSuffix(response, fmt.Sprintf("banned user %s", c.Username)) {
		return "", nil
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewBanUserAdminCommand(username string, reason string) *BanUserAdminCommand {
	return &BanUserAdminCommand{
		Username: username,
		Reason:   reason,
	}
}

type BanUserIDAdminCommand struct {
	SteamID string
}

func (c BanUserIDAdminCommand) ToCommand() string {
	return fmt.Sprintf("banid \"%s\"", c.SteamID)
}

func (c BanUserIDAdminCommand) ParseResponse(response string) (string, error) {
	if strings.Contains(strings.ToLower(response), fmt.Sprintf("banned steamid %s", strings.ToLower(c.SteamID))) {
		return "", nil
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewBanUserIDAdminCommand(steamID string) *BanUserIDAdminCommand {
	return &BanUserIDAdminCommand{
		SteamID: steamID,
	}
}
