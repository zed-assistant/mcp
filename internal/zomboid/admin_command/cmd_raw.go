package admincommand

type RawAdminCommand struct {
	Cmd string
}

func (c *RawAdminCommand) ToCommand() string {
	return c.Cmd
}

func (c *RawAdminCommand) ParseResponse(response string) (string, error) {
	return response, nil
}
