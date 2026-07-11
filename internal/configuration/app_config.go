package configuration

type LoggerConfig struct {
	JsonFormat   bool `koanf:"json_format"`
	DisableColor bool `koanf:"disable_color"`
}

type ServerConfig struct {
	Port int `koanf:"port"`
}

type AppConfig struct {
	Logger LoggerConfig `koanf:"logger"`
	Server ServerConfig `koanf:"server"`
}
