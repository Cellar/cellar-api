package settings

import (
	"cellar/pkg/settings/cryptography"
	"github.com/spf13/viper"
	"strings"
)

var Key string = "CONFIGURATION"

type IConfiguration interface {
	App() IAppConfiguration
	Redis() IRedisConfiguration
	Encryption() cryptography.IEncryptionConfiguration
	Logging() ILoggingConfiguration
}

type Configuration struct {
	app        IAppConfiguration
	redis      IRedisConfiguration
	encryption cryptography.IEncryptionConfiguration
	logging    ILoggingConfiguration
}

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Configuration{
		app:        NewAppConfiguration(),
		redis:      NewRedisConfiguration(),
		encryption: cryptography.NewEncryptionConfiguration(),
		logging:    NewLoggingConfiguration(),
	}
}

func (config Configuration) App() IAppConfiguration {
	return config.app
}

func (config Configuration) Redis() IRedisConfiguration {
	return config.redis
}

func (config Configuration) Encryption() cryptography.IEncryptionConfiguration {
	return config.encryption
}

func (config Configuration) Logging() ILoggingConfiguration { return config.logging }
