package admincommand

import "fmt"

type ServerMessageAdminCommand struct {
	Message string
}

func (c ServerMessageAdminCommand) ToCommand() string {
	return fmt.Sprintf("servermsg \"%s\"", c.Message)
}

func (c ServerMessageAdminCommand) ParseResponse(response string) (string, error) {
	return "", nil
}

func NewServerMessageAdminCommand(message string) ServerMessageAdminCommand {
	return ServerMessageAdminCommand{
		Message: message,
	}
}
