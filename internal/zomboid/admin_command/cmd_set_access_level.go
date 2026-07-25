package admincommand

import (
	"fmt"
	"strings"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type SetAccessLevelAdminCommand struct {
	Username    string
	AccessLevel string
}

func (c SetAccessLevelAdminCommand) ToCommand() string {
	return fmt.Sprintf("setaccesslevel %q %q", c.Username, c.AccessLevel)
}

func (c SetAccessLevelAdminCommand) ParseResponse(response string) (string, error) {
	switch {
	case response == fmt.Sprintf("User %s is now %s", c.Username, c.AccessLevel):
		return "", nil
	case strings.HasPrefix(response, fmt.Sprintf("Access Level '%s' unknown", c.AccessLevel)):
		return "", &domainerror.DomainError{
			InternalMessage: response,
			PublicMessage:   response,
			InternalCode:    domainerror.InvalidInput,
		}
	default:
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewSetAccessLevelAdminCommand(username string, accessLevel string) *SetAccessLevelAdminCommand {
	return &SetAccessLevelAdminCommand{
		Username:    username,
		AccessLevel: accessLevel,
	}
}
