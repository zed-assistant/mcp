package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type KickUserAdminCommand struct {
	Username string
	Reason   string
}

func (c KickUserAdminCommand) ToCommand() string {
	command := fmt.Sprintf("kickuser \"%s\"", c.Username)
	if c.Reason != "" {
		command += fmt.Sprintf(" -r \"%s\"", c.Reason)
	}
	return command
}

func (c KickUserAdminCommand) ParseResponse(response string) (string, error) {
	if response == fmt.Sprintf("User %s kicked.", c.Username) {
		return "", nil
	} else if response == fmt.Sprintf("User %s doesn't exist.", c.Username) {
		return "", domainerror.DomainError{
			InternalMessage: fmt.Sprintf("Game user %s not found", c.Username),
			PublicMessage:   fmt.Sprintf("User %s not found", c.Username),
			InternalCode:    domainerror.NotFound,
		}
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewKickUserAdminCommand(username string, reason string) *KickUserAdminCommand {
	return &KickUserAdminCommand{
		Username: username,
		Reason:   reason,
	}
}
