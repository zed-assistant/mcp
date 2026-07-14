package serverconfig

import (
	"path/filepath"
	"regexp"

	"gopkg.in/ini.v1"
)

var leadingCommentPrefix = regexp.MustCompile(`(?m)^#? *`)

type ServerConfigManager struct{}

func NewServerConfigManager() *ServerConfigManager {
	return &ServerConfigManager{}
}

func getIniPath(instanceHomeDir string) string {
	return filepath.Join(instanceHomeDir, "Server", "servertest.ini")
}

func loadIni(instanceHomeDir string) (*ini.File, string, error) {
	iniPath := getIniPath(instanceHomeDir)
	cfg := ini.LoadOptions{
		IgnoreInlineComment: true,
	}
	iniFile, err := ini.LoadSources(cfg, iniPath)
	if err != nil {
		return nil, "", err
	}

	ini.PrettyFormat = false

	return iniFile, iniPath, nil
}

func stripCommentPrefix(s string) string {
	return leadingCommentPrefix.ReplaceAllString(s, "")
}

func readConfigAsEntriesMap(iniFile *ini.File) map[string]ConfigEntry {
	configEntries := make(map[string]ConfigEntry)
	section, _ := iniFile.GetSection(ini.DefaultSection)

	for _, key := range section.Keys() {
		configEntries[key.Name()] = ConfigEntry{
			Key:         key.Name(),
			Value:       key.Value(),
			Description: stripCommentPrefix(key.Comment),
		}
	}

	return configEntries
}
