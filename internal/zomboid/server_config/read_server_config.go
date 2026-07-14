package serverconfig

import "fmt"

func (m *ServerConfigManager) ReadServerConfig(instanceHomeDir string) (map[string]ConfigEntry, error) {
	iniFile, _, err := loadIni(instanceHomeDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to load ini file for reading: %w", err)
	}

	return readConfigAsEntriesMap(iniFile), nil
}
