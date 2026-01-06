package settings

import (
	"cellar/pkg/settings/cryptography"
	"cellar/pkg/settings/datastore"
	"strings"

	"github.com/spf13/viper"
)

var Key = "CONFIGURATION"

//go:generate mockgen -destination=../mocks/mock_configuration.go -package=mocks cellar/pkg/settings IConfiguration
type IConfiguration interface {
	App() IAppConfiguration
	Datastore() datastore.IDatastoreConfiguration
	Encryption() cryptography.IEncryptionConfiguration
	Logging() ILoggingConfiguration
	RateLimit() IRateLimitConfiguration
}

type Configuration struct {
	app        IAppConfiguration
	datastore  datastore.IDatastoreConfiguration
	encryption cryptography.IEncryptionConfiguration
	logging    ILoggingConfiguration
	rateLimit  IRateLimitConfiguration
}

func NewConfiguration() *Configuration {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Configuration{
		app:        NewAppConfiguration(),
		datastore:  datastore.NewDatastoreConfiguration(),
		encryption: cryptography.NewEncryptionConfiguration(),
		logging:    NewLoggingConfiguration(),
		rateLimit:  NewRateLimitConfiguration(),
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

func (config Configuration) RateLimit() IRateLimitConfiguration { return config.rateLimit }
