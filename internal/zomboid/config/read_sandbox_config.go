package config

import (
	"fmt"
	"strings"

	"github.com/zed-assistant/mcp/internal/configuration"
	stringcompare "github.com/zed-assistant/mcp/internal/string_compare"
)

// sandboxNodeToOutput converts a parsed sandbox node into its JSON-facing shape:
// a leaf becomes a ConfigEntry{Key,Value,Description}, a group becomes a nested
// map[string]any of the same shape. keyFilters (if non-nil) are matched against
// each leaf's full dotted path (e.g. "ZombieLore.Speed"); groups with no
// matching descendants are pruned from the result.
func sandboxNodeToOutput(node *sandboxNode, path string, keyFilters []string) (any, bool, error) {
	if node.Leaf != nil {
		if keyFilters != nil {
			matched, err := matchesAnySandboxFilter(path, keyFilters)
			if err != nil {
				return nil, false, err
			}
			if !matched {
				return nil, false, nil
			}
		}
		key := path
		if idx := strings.LastIndex(path, "."); idx >= 0 {
			key = path[idx+1:]
		}
		return ConfigEntry{
			Key:         key,
			Value:       node.Leaf.Value,
			Description: node.Leaf.Description,
			Type:        node.Leaf.Kind.jsonType(),
		}, true, nil
	}

	result := make(map[string]any, len(node.ChildOrder))
	for _, childName := range node.ChildOrder {
		child := node.Children[childName]
		childPath := childName
		if path != "" {
			childPath = path + "." + childName
		}
		val, ok, err := sandboxNodeToOutput(child, childPath, keyFilters)
		if err != nil {
			return nil, false, err
		}
		if ok {
			result[childName] = val
		}
	}
	if len(result) == 0 {
		return nil, false, nil
	}
	return result, true, nil
}

func matchesAnySandboxFilter(path string, filters []string) (bool, error) {
	for _, filter := range filters {
		ok, err := stringcompare.CompareWithWildcard(path, filter)
		if err != nil {
			return false, fmt.Errorf("error comparing key '%s' with filter '%s': %w", path, filter, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (m *ConfigManager) ReadSandboxConfig(instanceConfig *configuration.ZomboidInstanceConfig, keyFilters []string) (map[string]any, error) {
	root, _, _, _, err := loadSandboxFile(instanceConfig.HomeDir, instanceConfig.ServerName)
	if err != nil {
		return nil, fmt.Errorf("unable to load sandbox config for reading: %w", err)
	}

	result := make(map[string]any, len(root.ChildOrder))
	for _, childName := range root.ChildOrder {
		child := root.Children[childName]
		val, ok, err := sandboxNodeToOutput(child, childName, keyFilters)
		if err != nil {
			return nil, err
		}
		if ok {
			result[childName] = val
		}
	}
	return result, nil
}
