package admincommand

import "strings"

type ReloadOptionsAdminCommand struct{}

func (c *ReloadOptionsAdminCommand) ToCommand() string {
	return "reloadoptions"
}

func (c *ReloadOptionsAdminCommand) ParseResponse(response string) (string, error) {
	return strings.TrimSpace(response), nil
}
