package configuration

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

//go:embed default/default.yml
var defaultConfig []byte

const (
	envPrefix = "ZED_ASSISTANT_MCP_"
	delim     = "."
)

func Load(path string) (*AppConfig, error) {
	k := koanf.New(delim)

	if err := k.Load(rawbytes.Provider(defaultConfig), yaml.Parser()); err != nil {
		return nil, err
	}

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, err
	}

	if err := k.Load(env.Provider(envPrefix, delim, func(s string) string {
		s = strings.TrimPrefix(s, envPrefix)
		s = strings.ToLower(s)
		return strings.ReplaceAll(s, "__", delim)
	}), nil); err != nil {
		return nil, fmt.Errorf("loading env: %w", err)
	}

	var config AppConfig
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &config, nil
}
