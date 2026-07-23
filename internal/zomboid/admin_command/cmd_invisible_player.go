package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type InvisiblePlayerAdminCommand struct {
	Username string
	Enabled  bool
}

func (c InvisiblePlayerAdminCommand) ToCommand() string {
	return fmt.Sprintf("invisibleplayer \"%s\" -%t", c.Username, c.Enabled)
}

func (c InvisiblePlayerAdminCommand) ParseResponse(response string) (string, error) {
	if c.Enabled && response == fmt.Sprintf("User %s is now invisible.", c.Username) {
		return "", nil
	} else if !c.Enabled && response == fmt.Sprintf("User %s is no longer invisible.", c.Username) {
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

func NewInvisiblePlayerAdminCommand(username string, enabled bool) *InvisiblePlayerAdminCommand {
	return &InvisiblePlayerAdminCommand{
		Username: username,
		Enabled:  enabled,
	}
}
