package admincommand

import (
	"fmt"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

func toggleCommandString(commandName, username string, enabled bool) string {
	return fmt.Sprintf("%s %q -%t", commandName, username, enabled)
}

func parseToggleResponse(response, username string, enabled bool, enabledMsg, disabledMsg string) (string, error) {
	switch {
	case enabled && response == fmt.Sprintf(enabledMsg, username):
		return "", nil
	case !enabled && response == fmt.Sprintf(disabledMsg, username):
		return "", nil
	case response == fmt.Sprintf("User %s not found.", username):
		return "", &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("Game user %s not found", username),
			PublicMessage:   response,
			InternalCode:    domainerror.NotFound,
		}
	default:
		return "", fmt.Errorf("unexpected response: %s", response)
	}
}
