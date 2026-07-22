package gamedatabase

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/zed-assistant/mcp/internal/configuration"
	filesystem "github.com/zed-assistant/mcp/internal/file_system"
)

type GameDatabaseManager struct{}

func NewGameDatabaseManager() *GameDatabaseManager {
	return &GameDatabaseManager{}
}

func (m *GameDatabaseManager) GetGameDatabase(instanceConfig configuration.ZomboidInstanceConfig) (*sql.DB, func() error, error) {
	dbFile := filepath.Join(instanceConfig.HomeDir, "db", instanceConfig.ServerName+".db")
	dbFileExists, err := filesystem.FileExists(dbFile)
	if err != nil {
		return nil, nil, err
	}
	if !dbFileExists {
		return nil, nil, fmt.Errorf("game database file does not exist: %s", dbFile)
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbFile))
	if err != nil {
		return nil, nil, err
	}

	return db, db.Close, nil
}
