package configuration

import "time"

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
	PendingAuthTTL    time.Duration   `koanf:"pending_auth_ttl"`
	IDP               OAuth2IDPConfig `koanf:"idp"`
}

type OAuth2IDPConfig struct {
	Type  string                `koanf:"type"`
	Local *OAuth2IDPLocalConfig `koanf:"local"`
}

type LocalUserConfig struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

type OAuth2IDPLocalConfig struct {
	Users []LocalUserConfig `koanf:"users"`
}

const defaultServerName = "servertest"

type ZomboidInstanceConfig struct {
	Name       string   `koanf:"name"`
	HomeDir    string   `koanf:"home_dir"`
	ServerName string   `koanf:"server_name"`
	RCONHost   string   `koanf:"rcon_host"`
	Users      []string `koanf:"users"`
}

type ZomboidConfig struct {
	Instances map[string]ZomboidInstanceConfig `koanf:"instances"`
}

type AppConfig struct {
	Logger  LoggerConfig  `koanf:"logger"`
	Server  ServerConfig  `koanf:"server"`
	OAuth2  OAuth2Config  `koanf:"oauth2"`
	Zomboid ZomboidConfig `koanf:"zomboid"`
}

func (c *AppConfig) applyDefaults() {
	for id, instance := range c.Zomboid.Instances {
		if instance.ServerName == "" {
			instance.ServerName = defaultServerName
			c.Zomboid.Instances[id] = instance
		}
	}
}
