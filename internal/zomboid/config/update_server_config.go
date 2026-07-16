package config

import (
	"fmt"
	"strings"

	"github.com/zed-assistant/mcp/internal/configuration"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
	"gopkg.in/ini.v1"
)

func NewInvalidKeysError(invalidKeys []string) *domainerror.DomainError {
	keys := strings.Join(invalidKeys, ", ")
	return &domainerror.DomainError{
		InternalMessage: "invalid config keys: " + keys,
		PublicMessage:   "invalid config keys: " + keys,
		InternalCode:    domainerror.InvalidInput,
	}
}

func (m *ConfigManager) UpdateServerConfig(instanceConfig *configuration.ZomboidInstanceConfig, newConfig map[string]string) error {
	iniFile, iniPath, err := loadIni(instanceConfig.HomeDir, instanceConfig.ServerName)
	if err != nil {
		return fmt.Errorf("failed to load ini file for update: %w", err)
	}

	invalidKeys := []string{}

	section, _ := iniFile.GetSection(ini.DefaultSection)

	for key := range newConfig {
		if !section.HasKey(key) {
			invalidKeys = append(invalidKeys, key)
			continue
		}
	}

	if len(invalidKeys) > 0 {
		return NewInvalidKeysError(invalidKeys)
	}

	for key, value := range newConfig {
		section.Key(key).SetValue(value)
	}

	if err := iniFile.SaveTo(iniPath); err != nil {
		return fmt.Errorf("failed to save ini file: %w", err)
	}

	return nil
}
