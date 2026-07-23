package admincommand

import (
	"fmt"
	"strings"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type AddXPAdminCommand struct {
	Username string
	Perk     string
	Amount   int
}

func (c AddXPAdminCommand) ToCommand() string {
	return fmt.Sprintf("addxp \"%s\" %s=%d", c.Username, c.Perk, c.Amount)
}

func (c AddXPAdminCommand) ParseResponse(response string) (string, error) {
	if response == fmt.Sprintf("Added %d.0 %s xp's to %s", c.Amount, c.Perk, c.Username) {
		return "", nil
	} else if response == "No such user" {
		return "", &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("Game user %s not found", c.Username),
			PublicMessage:   fmt.Sprintf("User %s not found", c.Username),
			InternalCode:    domainerror.NotFound,
		}
	} else if strings.Contains(response, "List of available perks") {
		return "", &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("Unknown perk %s. Server response: %s", c.Perk, response),
			PublicMessage:   fmt.Sprintf("Unknown perk %s. %s", c.Perk, response),
			InternalCode:    domainerror.InvalidInput,
		}
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewAddXPAdminCommand(username string, perk string, amount int) *AddXPAdminCommand {
	return &AddXPAdminCommand{
		Username: username,
		Perk:     perk,
		Amount:   amount,
	}
}
