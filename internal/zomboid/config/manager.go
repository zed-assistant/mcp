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
