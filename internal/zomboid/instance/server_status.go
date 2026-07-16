package instance

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
)

type ServerStatusPlayer struct {
	Username    string `json:"username" jsonschema:"The username of the player"`
	AccessLevel string `json:"accessLevel" jsonschema:"The access level of the player"`
}

type ServerStatus struct {
	Online      bool                 `json:"online" jsonschema:"Whether the server is online"`
	Players     []ServerStatusPlayer `json:"players" jsonschema:"The list of players currently online"`
	PlayerCount int                  `json:"playerCount" jsonschema:"The number of players currently online"`
	MaxPlayers  int                  `json:"maxPlayers" jsonschema:"The maximum number of players allowed on the server"`
}

type GetServerStatusInput struct {
	InstanceID string
}

func (m *ZomboidInstanceManager) GetServerStatus(ctx context.Context, principal authorization.Principal, input *GetServerStatusInput) (*ServerStatus, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	m.instanceLockManager.RLock(input.InstanceID)
	defer m.instanceLockManager.RUnlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]

	isOnline := true

	players, err := admincommand.ExecuteSingleAdminCommand(m.adminCommandManager, &instanceCfg, admincommand.NewListPlayersAdminCommand())
	if err != nil {

		var opErr *net.OpError
		if errors.As(err, &opErr) && opErr.Op == "dial" {
			isOnline = false
		} else {
			return nil, fmt.Errorf("failed to get players list: %w", err)
		}
	}

	playersList := make([]ServerStatusPlayer, 0, len(players))
	if isOnline {
		for _, player := range players {
			playersList = append(playersList, ServerStatusPlayer{
				Username: player,
			})
		}
	}

	cfg, err := m.serverConfigManager.ReadServerConfig(&instanceCfg, []string{"MaxPlayers"})
	if err != nil {
		return nil, fmt.Errorf("failed to read server config: %w", err)
	}

	maxPlayers, ok := cfg["MaxPlayers"]
	if !ok {
		return nil, fmt.Errorf("failed to read MaxPlayers from server config")
	}
	numberOfMaxPlayers, err := strconv.Atoi(maxPlayers.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MaxPlayers: %w", err)
	}

	return &ServerStatus{
		Online:      isOnline,
		Players:     playersList,
		PlayerCount: len(players),
		MaxPlayers:  numberOfMaxPlayers,
	}, nil
}
