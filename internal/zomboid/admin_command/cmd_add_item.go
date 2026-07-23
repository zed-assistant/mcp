package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

type AddItemAdminCommand struct {
	Username string
	ItemType string
	Count    int
}

func (c AddItemAdminCommand) ToCommand() string {
	return fmt.Sprintf("additem \"%s\" \"%s\" %d", c.Username, c.ItemType, c.Count)
}

func (c AddItemAdminCommand) ParseResponse(response string) (string, error) {
	if response == fmt.Sprintf("Item %s Added in %s's inventory.", c.ItemType, c.Username) {
		return "", nil
	} else if response == "No such user" {
		return "", &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("Game user %s not found", c.Username),
			PublicMessage:   fmt.Sprintf("User %s not found", c.Username),
			InternalCode:    domainerror.NotFound,
		}
	} else if response == fmt.Sprintf("Item %s doesn't exist.", c.ItemType) {
		return "", &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("Item %s doesn't exist", c.ItemType),
			PublicMessage:   response,
			InternalCode:    domainerror.InvalidInput,
		}
	} else {
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}

func NewAddItemAdminCommand(username string, itemType string, count int) *AddItemAdminCommand {
	if count <= 0 {
		count = 1
	}
	return &AddItemAdminCommand{
		Username: username,
		ItemType: itemType,
		Count:    count,
	}
}
