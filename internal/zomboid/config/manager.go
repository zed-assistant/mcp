package config

import (
	"path/filepath"
	"regexp"

	"gopkg.in/ini.v1"
)

var leadingCommentPrefix = regexp.MustCompile(`(?m)^#? *`)

type ServerConfigManager struct{}

func NewConfigManager() *ServerConfigManager {
	return &ServerConfigManager{}
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
