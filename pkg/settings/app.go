package settings

import (
	"github.com/spf13/viper"
)

type IAppConfiguration interface {
	ClientAddress() string
	BindAddress() string
}

const (
	appKey           = "app."
	appClientAddress = appKey + "client_address"
	appBindAddress   = appKey + "bind_address"
)

type AppConfiguration struct{}

func NewAppConfiguration() *AppConfiguration {
	defaultAddress := "127.0.0.1:8080"
	viper.SetDefault(appBindAddress, defaultAddress)
	viper.SetDefault(appClientAddress, "http://"+defaultAddress)
	return &AppConfiguration{}
}

func (app AppConfiguration) ClientAddress() string {
	return viper.GetString(appClientAddress)
}

func (app AppConfiguration) BindAddress() string {
	return viper.GetString(appBindAddress)
}
