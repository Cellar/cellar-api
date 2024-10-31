package settings

import (
	"cellar/pkg/settings/cryptography"
	"cellar/pkg/settings/datastore"
	"github.com/spf13/viper"
	"strings"
)

var Key string = "CONFIGURATION"

type IConfiguration interface {
	App() IAppConfiguration
	Datastore() datastore.IDatastoreConfiguration
	Encryption() cryptography.IEncryptionConfiguration
	Logging() ILoggingConfiguration
}

type Configuration struct {
	app        IAppConfiguration
	datastore  datastore.IDatastoreConfiguration
	encryption cryptography.IEncryptionConfiguration
	logging    ILoggingConfiguration
}

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Configuration{
		app:        NewAppConfiguration(),
		datastore:  datastore.NewDatastoreConfiguration(),
		encryption: cryptography.NewEncryptionConfiguration(),
		logging:    NewLoggingConfiguration(),
	}
}

func (config Configuration) App() IAppConfiguration {
	return config.app
}

func (config Configuration) Datastore() datastore.IDatastoreConfiguration { return config.datastore }

func (config Configuration) Encryption() cryptography.IEncryptionConfiguration {
	return config.encryption
}

func (config Configuration) Logging() ILoggingConfiguration { return config.logging }
