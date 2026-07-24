package mcptool

import (
	"testing"

	"github.com/zed-assistant/mcp/internal/zomboid/config"
)

func TestReadConfigInput_ValidatesConfigType(t *testing.T) {
	valid := []config.ConfigType{config.ConfigTypeServer, config.ConfigTypeSandbox}
	for _, ct := range valid {
		in := ReadConfigInput{InstanceId: "abc", ConfigType: ct}
		if err := validate.Struct(in); err != nil {
			t.Errorf("expected configType %q to be valid, got error: %v", ct, err)
		}
	}

	bad := ReadConfigInput{InstanceId: "abc", ConfigType: "bogus"}
	if err := validate.Struct(bad); err == nil {
		t.Errorf("expected an invalid configType to fail validation")
	}
}

func TestUpdateConfigInput_ValidatesConfigTypeAndAllowsNestedUpdates(t *testing.T) {
	in := UpdateConfigInput{
		InstanceId: "abc",
		ConfigType: config.ConfigTypeSandbox,
		Updates: map[string]any{
			"Zombies": "4",
			"ZombieLore": map[string]any{
				"Speed": "2",
			},
		},
	}
	if err := validate.Struct(in); err != nil {
		t.Errorf("expected nested updates with valid configType to pass validation, got: %v", err)
	}

	bad := UpdateConfigInput{InstanceId: "abc", ConfigType: "bogus", Updates: map[string]any{}}
	if err := validate.Struct(bad); err == nil {
		t.Errorf("expected an invalid configType to fail validation")
	}
}
