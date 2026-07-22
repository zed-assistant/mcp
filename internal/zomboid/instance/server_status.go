package instance

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
)

type ServerStatusPlayer struct {
	Username    string     `json:"username" jsonschema:"The username of the player"`
	Role        string     `json:"role,omitempty" jsonschema:"The role of the player"`
	ConnectedAt *time.Time `json:"connectedAt,omitempty" jsonschema:"The time the player connected to the server"`
	SteamID     string     `json:"steamId,omitempty" jsonschema:"The Steam ID of the player"`
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
		playersList, err = m.mapOnlinePlayersToServerStatusPlayers(ctx, instanceCfg, players)
		if err != nil {
			return nil, fmt.Errorf("failed to map online players to server status players: %w", err)
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

func (m *ZomboidInstanceManager) mapOnlinePlayersToServerStatusPlayers(ctx context.Context, instanceConfig configuration.ZomboidInstanceConfig, players []string) ([]ServerStatusPlayer, error) {
	if len(players) == 0 {
		return []ServerStatusPlayer{}, nil
	}

	allUsers, err := m.whitelistManager.GetAllUsers(ctx, instanceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	index := make(map[string]int, len(allUsers))
	for i := range allUsers {
		index[allUsers[i].Username] = i
	}

	results := make([]ServerStatusPlayer, 0, len(players))
	for _, name := range players {
		r := ServerStatusPlayer{Username: name}
		if i, ok := index[name]; ok {
			u := allUsers[i]
			r.Role = u.RoleName
			r.ConnectedAt = u.LastConnection
			r.SteamID = u.SteamID
		}
		results = append(results, r)
	}

	return results, nil

}
