package settings

import (
	"github.com/spf13/viper"
	"strings"
)

var Key string = "CONFIGURATION"

type IConfiguration interface {
	App() IAppConfiguration
	Redis() IRedisConfiguration
	Vault() IVaultConfiguration
	Logging() ILoggingConfiguration
}

type Configuration struct {
	app     IAppConfiguration
	redis   IRedisConfiguration
	vault   IVaultConfiguration
	logging ILoggingConfiguration
}

var configuration IConfiguration

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Configuration{
		app:   NewAppConfiguration(),
		redis: NewRedisConfiguration(),
		vault: NewVaultConfiguration(),
		logging: NewLoggingConfiguration(),
	}
}
func (config Configuration) App() IAppConfiguration {
	return config.app
}

func (config Configuration) Redis() IRedisConfiguration {
	return config.redis
}

func (config Configuration) Vault() IVaultConfiguration {
	return config.vault
}

func (config Configuration) Logging() ILoggingConfiguration {
	return config.logging
}
