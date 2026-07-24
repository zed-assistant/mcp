package config

import (
	"fmt"
	"strings"

	"github.com/zed-assistant/mcp/internal/configuration"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
	"gopkg.in/ini.v1"
)

// FlattenServerConfigUpdates validates a (possibly nested) update tree for
// configType=server: ini config has no groups, so every value must be a plain
// JSON string. It collects every problem before returning, matching the
// reject-before-applying-anything behavior used for sandbox config.
func FlattenServerConfigUpdates(updates map[string]any) (map[string]string, error) {
	flat := make(map[string]string, len(updates))
	var problems []string

	for key, rawVal := range updates {
		strVal, ok := rawVal.(string)
		if !ok {
			if _, isNested := rawVal.(map[string]any); isNested {
				problems = append(problems, fmt.Sprintf("'%s': nested values are not supported for server config", key))
			} else {
				problems = append(problems, fmt.Sprintf("'%s': value must be provided as a JSON string, got %s", key, jsonTypeName(rawVal)))
			}
			continue
		}
		flat[key] = strVal
	}

	if len(problems) > 0 {
		return nil, NewInvalidConfigUpdateError(problems)
	}
	return flat, nil
}

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
