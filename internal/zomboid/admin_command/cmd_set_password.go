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
	return fmt.Sprintf("setpassword %q %q", c.Username, c.Password)
}

func (c SetPasswordAdminCommand) ParseResponse(response string) (string, error) {
	const prefix = "Your new password is "
	if rest, ok := strings.CutPrefix(response, prefix); ok {
		return rest, nil
	}
	return "", fmt.Errorf("unexpected response: %s", response)
}

func NewSetPasswordAdminCommand(username string, password string) *SetPasswordAdminCommand {
	return &SetPasswordAdminCommand{
		Username: username,
		Password: password,
	}
}
