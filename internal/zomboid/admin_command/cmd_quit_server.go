package admincommand

import "fmt"

type QuitServerAdminCommand struct{}

func (c QuitServerAdminCommand) ToCommand() string {
	return "quit"
}

func (c QuitServerAdminCommand) ParseResponse(response string) (string, error) {
	if response == "Quit" {
		return "", nil
	}
	return "", fmt.Errorf("unexpected response: %s", response)
}

func NewQuitServerAdminCommand() *QuitServerAdminCommand {
	return &QuitServerAdminCommand{}
}
