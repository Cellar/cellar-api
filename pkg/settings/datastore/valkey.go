package datastore

import (
	"github.com/spf13/viper"
)

type IValkeyConfiguration interface {
	Enabled() bool
	Host() string
	Port() int
	Username() string
	Password() string
	DB() int
	TLSEnabled() bool
}

const (
	valkeyKey           = datastoreKey + "valkey."
	valkeyEnabledKey    = valkeyKey + "enabled"
	valkeyHostKey       = valkeyKey + "host"
	valkeyPortKey       = valkeyKey + "port"
	valkeyUsernameKey   = valkeyKey + "username"
	valkeyPasswordKey   = valkeyKey + "password"
	valkeyDBKey         = valkeyKey + "db"
	valkeyTLSEnabledKey = valkeyKey + "tls_enabled"
)

type ValkeyConfiguration struct{}

func NewValkeyConfiguration() *ValkeyConfiguration {
	viper.SetDefault(valkeyHostKey, "localhost")
	viper.SetDefault(valkeyPortKey, 6379)
	viper.SetDefault(valkeyUsernameKey, "")
	viper.SetDefault(valkeyPasswordKey, "")
	viper.SetDefault(valkeyDBKey, 0)
	return &ValkeyConfiguration{}
}

func (vk ValkeyConfiguration) Enabled() bool {
	return viper.GetBool(valkeyEnabledKey)
}

func (vk ValkeyConfiguration) Host() string {
	return viper.GetString(valkeyHostKey)
}

func (vk ValkeyConfiguration) Port() int {
	return viper.GetInt(valkeyPortKey)
}

func (vk ValkeyConfiguration) Username() string {
	return viper.GetString(valkeyUsernameKey)
}

func (vk ValkeyConfiguration) Password() string {
	return viper.GetString(valkeyPasswordKey)
}

func (vk ValkeyConfiguration) DB() int {
	return viper.GetInt(valkeyDBKey)
}

func (vk ValkeyConfiguration) TLSEnabled() bool {
	return viper.GetBool(valkeyTLSEnabledKey)
}
