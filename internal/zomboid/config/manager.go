package config

import (
	"fmt"
	"os"
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

func getSandboxLuaPath(instanceHomeDir string, serverName string) string {
	return filepath.Join(instanceHomeDir, "Server", serverName+"_SandboxVars.lua")
}

func loadSandboxFile(instanceHomeDir string, serverName string) (root *sandboxNode, path string, src []byte, mode os.FileMode, err error) {
	path = getSandboxLuaPath(instanceHomeDir, serverName)

	info, err := os.Stat(path)
	if err != nil {
		return nil, "", nil, 0, fmt.Errorf("unable to stat sandbox lua file: %w", err)
	}

	src, err = os.ReadFile(path) // nolint:gosec
	if err != nil {
		return nil, "", nil, 0, fmt.Errorf("unable to read sandbox lua file: %w", err)
	}

	root, err = parseSandboxFile(src)
	if err != nil {
		return nil, "", nil, 0, fmt.Errorf("unable to parse sandbox lua file %s: %w", path, err)
	}

	return root, path, src, info.Mode(), nil
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
