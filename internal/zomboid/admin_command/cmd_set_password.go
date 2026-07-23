package admincommand

import (
	"fmt"
	"strings"
)

type SetPasswordAdminCommand struct {
	Username string
	Password string
}

func (c SetPasswordAdminCommand) ToCommand() string {
	return fmt.Sprintf("setpassword \"%s\" \"%s\"", c.Username, c.Password)
}

func (c SetPasswordAdminCommand) ParseResponse(response string) (string, error) {
	const prefix = "Your new password is "
	if strings.HasPrefix(response, prefix) {
		return strings.TrimPrefix(response, prefix), nil
	}
	return "", fmt.Errorf("unexpected response: %s", response)
}

func NewSetPasswordAdminCommand(username string, password string) *SetPasswordAdminCommand {
	return &SetPasswordAdminCommand{
		Username: username,
		Password: password,
	}
}
