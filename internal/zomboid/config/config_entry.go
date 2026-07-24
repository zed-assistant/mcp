package config

type ConfigEntry struct {
	Key         string `json:"key" jsonschema:"Name of the configuration entry"`
	Value       string `json:"value,omitempty" jsonschema:"Value of the configuration entry"`
	Description string `json:"description,omitempty" jsonschema:"Description of the configuration entry. May contain information about the expected value, default value, and any other relevant details."`
	Type        string `json:"type,omitempty" jsonschema:"The kind of value this entry holds: 'integer', 'number' (a float), 'boolean', or 'string'. On update, the value must still be sent as a JSON string, but its content must match this kind (type='integer' needs a whole number like \"4\", no decimal point; type='number' needs a numeric string and may include a decimal point, e.g. \"1.5\" or \"1\"; type='boolean' needs exactly \"true\" or \"false\"). An integer field can never be updated with a fractional value. Not set for configType='server', where every value is always a plain string."`
}

type ConfigType string

const (
	ConfigTypeServer  ConfigType = "server"
	ConfigTypeSandbox ConfigType = "sandbox"
)
