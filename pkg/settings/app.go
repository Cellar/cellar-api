package settings

import (
	"github.com/spf13/viper"
)

type IAppConfiguration interface {
	BindAddress() string
	ClientAddress() string
	Version() string
	MaxFileSizeMB() int
}

const (
	appKey              = "app."
	appVersionKey       = appKey + "version"
	appClientAddressKey = appKey + "client_address"
	appBindAddressKey   = appKey + "bind_address"
	appMaxFileSizeMBKey = appKey + "maxFileSizeMB"
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
	return viper.GetInt(appMaxFileSizeMBKey)
}
