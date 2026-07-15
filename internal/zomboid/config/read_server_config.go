package config

import (
	"fmt"

	stringcompare "github.com/zed-assistant/mcp/internal/string_compare"
	"gopkg.in/ini.v1"
)

func stripCommentPrefix(s string) string {
	return leadingCommentPrefix.ReplaceAllString(s, "")
}

func readServerConfigAsEntriesMap(iniFile *ini.File, keyFilters []string) (map[string]ConfigEntry, error) {
	configEntries := make(map[string]ConfigEntry)
	section, _ := iniFile.GetSection(ini.DefaultSection)

	for _, key := range section.Keys() {

		if keyFilters != nil {
			matched := false
			for _, filter := range keyFilters {
				ok, err := stringcompare.CompareWithWildcard(key.Name(), filter)
				if err != nil {
					return nil, fmt.Errorf("Error comparing key '%s' with filter '%s': %w", key.Name(), filter, err)
				}
				if ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		configEntries[key.Name()] = ConfigEntry{
			Key:         key.Name(),
			Value:       key.Value(),
			Description: stripCommentPrefix(key.Comment),
		}
	}

	return configEntries, nil
}

func (m *ServerConfigManager) ReadServerConfig(instanceHomeDir string, keysFilter []string) (map[string]ConfigEntry, error) {
	iniFile, _, err := loadIni(instanceHomeDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to load ini file for reading: %w", err)
	}

	return readServerConfigAsEntriesMap(iniFile, keysFilter)
}
