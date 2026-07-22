package whitelist

import (
	"context"
	"database/sql"
	"time"

	"github.com/zed-assistant/mcp/internal/configuration"
)

func (m *WhitelistManager) GetAllUsers(ctx context.Context, instanceConfig configuration.ZomboidInstanceConfig) ([]*User, error) {
	db, closeDb, err := m.gameDb.GetGameDatabase(instanceConfig)
	if err != nil {
		return nil, err
	}
	defer func() { _ = closeDb() }()

	rows, err := db.QueryContext(ctx, "SELECT w.id, w.world, w.username, w.lastConnection, w.role, r.name AS roleName, w.steamid FROM whitelist AS w LEFT JOIN role AS r ON w.role = r.id;")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var users []*User
	for rows.Next() {
		var user User
		var lastConnectionValue sql.NullString
		var steamIDValue sql.NullString

		err := rows.Scan(&user.ID, &user.World, &user.Username, &lastConnectionValue, &user.RoleID, &user.RoleName, &steamIDValue)
		if err != nil {
			return nil, err
		}
		if lastConnectionValue.Valid && lastConnectionValue.String != "" {
			lastConnection, err := time.Parse("2006-01-02 15:04:05", lastConnectionValue.String)
			if err != nil {
				return nil, err
			}
			user.LastConnection = &lastConnection
		}
		if steamIDValue.Valid && steamIDValue.String != "" {
			user.SteamID = steamIDValue.String
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
