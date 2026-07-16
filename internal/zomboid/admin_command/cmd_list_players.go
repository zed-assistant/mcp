package admincommand

import "strings"

type ListPlayersAdminCommand struct{}

func (c ListPlayersAdminCommand) ToCommand() string {
	return "players"
}

func (c ListPlayersAdminCommand) ParseResponse(response string) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	players := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") {
			player := strings.TrimSpace(strings.TrimPrefix(line, "-"))
			players = append(players, player)
		}
	}
	return players, nil
}

func NewListPlayersAdminCommand() ListPlayersAdminCommand {
	return ListPlayersAdminCommand{}
}
