package admincommand

import "fmt"

type ReloadAllLuaAdminCommand struct{}

func (c ReloadAllLuaAdminCommand) ToCommand() string {
	return "reloadalllua"
}

func (c ReloadAllLuaAdminCommand) ParseResponse(response string) (string, error) {
	if response == "Lua files reloaded" {
		return "", nil
	}
	return "", fmt.Errorf("unexpected response: %s", response)
}

func NewReloadAllLuaAdminCommand() *ReloadAllLuaAdminCommand {
	return &ReloadAllLuaAdminCommand{}
}
