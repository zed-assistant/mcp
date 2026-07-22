package whitelist

import (
	"database/sql"

	"github.com/zed-assistant/mcp/internal/configuration"
)

type gameDb interface {
	GetGameDatabase(instanceConfig configuration.ZomboidInstanceConfig) (*sql.DB, func() error, error)
}

type WhitelistManager struct {
	gameDb gameDb
}

func NewWhitelistManager(gameDb gameDb) *WhitelistManager {
	return &WhitelistManager{
		gameDb: gameDb,
	}
}
