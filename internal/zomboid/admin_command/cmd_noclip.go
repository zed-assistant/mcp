package admincommand

type NoclipAdminCommand struct {
	Username string
	Enabled  bool
}

func (c NoclipAdminCommand) ToCommand() string {
	return toggleCommandString("noclip", c.Username, c.Enabled)
}

func (c NoclipAdminCommand) ParseResponse(response string) (string, error) {
	return parseToggleResponse(response, c.Username, c.Enabled, "User %s won't collide.", "User %s will collide.")
}

func NewNoclipAdminCommand(username string, enabled bool) *NoclipAdminCommand {
	return &NoclipAdminCommand{
		Username: username,
		Enabled:  enabled,
	}
}
