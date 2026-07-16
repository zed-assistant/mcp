package config

import (
	"path/filepath"
	"regexp"

	"gopkg.in/ini.v1"
)

var leadingCommentPrefix = regexp.MustCompile(`(?m)^#? *`)

type ConfigManager struct{}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

func getIniPath(instanceHomeDir string, serverName string) string {
	return filepath.Join(instanceHomeDir, "Server", serverName+".ini")
}

func loadIni(instanceHomeDir string, serverName string) (*ini.File, string, error) {
	iniPath := getIniPath(instanceHomeDir, serverName)
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
