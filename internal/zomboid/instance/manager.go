package instance

import "github.com/zed-assistant/mcp/internal/configuration"

type ZomboidInstanceManager struct {
	appConfig *configuration.AppConfig
}

func NewZomboidInstanceManager(appConfig *configuration.AppConfig) *ZomboidInstanceManager {
	return &ZomboidInstanceManager{
		appConfig: appConfig,
	}
}
