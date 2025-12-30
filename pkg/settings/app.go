package settings

import (
	"github.com/spf13/viper"
)

//go:generate mockgen -destination=../mocks/mock_app_configuration.go -package=mocks cellar/pkg/settings IAppConfiguration
type IAppConfiguration interface {
	BindAddress() string
	ClientAddress() string
	Version() string
	MaxFileSizeMB() int
	MaxAccessCount() int
	MaxExpirationSeconds() int
}

const (
	appKey                     = "app."
	appVersionKey              = appKey + "version"
	appClientAddressKey        = appKey + "client_address"
	appBindAddressKey          = appKey + "bind_address"
	appMaxFileSizeMBKey        = appKey + "max_file_size_mb"
	appMaxAccessCountKey       = appKey + "max_access_count"
	appMaxExpirationSecondsKey = appKey + "max_expiration_seconds"
)

var version string

func SetAppVersion(appVersion string) {
	version = appVersion
}

type AppConfiguration struct{}

func NewAppConfiguration() *AppConfiguration {
	defaultAddress := "127.0.0.1:8080"
	viper.Set(appVersionKey, version)
	viper.SetDefault(appBindAddressKey, defaultAddress)
	viper.SetDefault(appClientAddressKey, "http://"+defaultAddress)
	viper.SetDefault(appMaxFileSizeMBKey, 8)
	viper.SetDefault(appMaxAccessCountKey, 100)
	viper.SetDefault(appMaxExpirationSecondsKey, 604800)
	return &AppConfiguration{}
}

func (app AppConfiguration) BindAddress() string {
	return viper.GetString(appBindAddressKey)
}

func (app AppConfiguration) ClientAddress() string {
	return viper.GetString(appClientAddressKey)
}

func (app AppConfiguration) Version() string {
	return viper.GetString(appVersionKey)
}

func (app AppConfiguration) MaxFileSizeMB() int {
	value := viper.GetInt(appMaxFileSizeMBKey)
	if value < 0 {
		return 0
	}
	return value
}

func (app AppConfiguration) MaxAccessCount() int {
	value := viper.GetInt(appMaxAccessCountKey)
	if value < 1 {
		return 1
	}
	return value
}

func (app AppConfiguration) MaxExpirationSeconds() int {
	value := viper.GetInt(appMaxExpirationSecondsKey)
	if value < 900 {
		return 900
	}
	return value
}
