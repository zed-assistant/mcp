package admincommand

type InvisiblePlayerAdminCommand struct {
	Username string
	Enabled  bool
}

func (c InvisiblePlayerAdminCommand) ToCommand() string {
	return toggleCommandString("invisibleplayer", c.Username, c.Enabled)
}

func (c InvisiblePlayerAdminCommand) ParseResponse(response string) (string, error) {
	return parseToggleResponse(response, c.Username, c.Enabled, "User %s is now invisible.", "User %s is no longer invisible.")
}

func NewInvisiblePlayerAdminCommand(username string, enabled bool) *InvisiblePlayerAdminCommand {
	return &InvisiblePlayerAdminCommand{
		Username: username,
		Enabled:  enabled,
	}
}
