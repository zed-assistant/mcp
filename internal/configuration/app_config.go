package configuration

type LoggerConfig struct {
	JsonFormat   bool `koanf:"json_format"`
	DisableColor bool `koanf:"disable_color"`
}

type ServerConfig struct {
	Port        int    `koanf:"port"`
	ExternalUrl string `koanf:"external_url"`
}

type OAuth2Config struct {
	SigningSecret     B64EncodedBytes `koanf:"signing_secret"`
	IdTokenSigningKey *RSAPrivateKey  `koanf:"id_token_signing_key"`
}

type AppConfig struct {
	Logger LoggerConfig `koanf:"logger"`
	Server ServerConfig `koanf:"server"`
	OAuth2 OAuth2Config `koanf:"oauth2"`
}
