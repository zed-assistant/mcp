package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type AddUserAdminCommand struct {
	Username string
	Password string
}

func (c AddUserAdminCommand) ToCommand() string {
	return fmt.Sprintf("adduser %q %q", c.Username, c.Password)
}

func (c AddUserAdminCommand) ParseResponse(response string) (string, error) {
	switch response {
	case fmt.Sprintf("User %s created with password", c.Username):
		return "", nil
	case "A user with this name already exists":
		return "", &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("A user with the name %s already exists", c.Username),
			PublicMessage:   response,
			InternalCode:    domainerror.Conflict,
		}
	default:
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewAddUserAdminCommand(username string, password string) *AddUserAdminCommand {
	return &AddUserAdminCommand{
		Username: username,
		Password: password,
	}
}
