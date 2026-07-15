package config

type ConfigEntry struct {
	Key         string `json:"key" jsonschema:"Name of the configuration entry"`
	Value       string `json:"value,omitempty" jsonschema:"Value of the configuration entry"`
	Description string `json:"description,omitempty" jsonschema:"Description of the configuration entry. May contain information about the expected value, default value, and any other relevant details."`
}
