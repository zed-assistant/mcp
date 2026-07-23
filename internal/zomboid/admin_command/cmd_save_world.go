package admincommand

import "fmt"

type SaveWorldAdminCommand struct{}

func (c SaveWorldAdminCommand) ToCommand() string {
	return "save"
}

func (c SaveWorldAdminCommand) ParseResponse(response string) (string, error) {
	if response == "World saved" {
		return "", nil
	}
	return "", fmt.Errorf("unexpected response: %s", response)
}

func NewSaveWorldAdminCommand() *SaveWorldAdminCommand {
	return &SaveWorldAdminCommand{}
}
