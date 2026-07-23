package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type NoclipAdminCommand struct {
	Username string
	Enabled  bool
}

func (c NoclipAdminCommand) ToCommand() string {
	return fmt.Sprintf("noclip \"%s\" -%t", c.Username, c.Enabled)
}

func (c NoclipAdminCommand) ParseResponse(response string) (string, error) {
	if c.Enabled && response == fmt.Sprintf("User %s won't collide.", c.Username) {
		return "", nil
	} else if !c.Enabled && response == fmt.Sprintf("User %s will collide.", c.Username) {
		return "", nil
	} else if response == fmt.Sprintf("User %s not found.", c.Username) {
		return "", &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("Game user %s not found", c.Username),
			PublicMessage:   response,
			InternalCode:    domainerror.NotFound,
		}
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewNoclipAdminCommand(username string, enabled bool) *NoclipAdminCommand {
	return &NoclipAdminCommand{
		Username: username,
		Enabled:  enabled,
	}
}
