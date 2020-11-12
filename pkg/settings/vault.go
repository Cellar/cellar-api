package settings

import (
	"github.com/spf13/viper"
)

type IVaultConfiguration interface {
	Address() string
	RoleID() string
	SecretID() string
	TokenName() string
}

const (
	vaultKey          = "vault."
	vaultAddressKey   = vaultKey + "address"
	vaultRoleIDKey    = vaultKey + "role_id"
	vaultSecretIDKey  = vaultKey + "secret_id"
	vaultTokenNameKey = vaultKey + "token_name"
)

type VaultConfiguration struct{}

func NewVaultConfiguration() *VaultConfiguration {
	viper.SetDefault(vaultAddressKey, "http://localhost:8200")
	viper.SetDefault(vaultRoleIDKey, "")
	viper.SetDefault(vaultSecretIDKey, "")
	viper.SetDefault(vaultTokenNameKey, "")
	return &VaultConfiguration{}
}

func (vlt VaultConfiguration) Address() string {
	return viper.GetString(vaultAddressKey)
}

func (vlt VaultConfiguration) RoleID() string {
	return viper.GetString(vaultRoleIDKey)
}

func (vlt VaultConfiguration) SecretID() string {
	return viper.GetString(vaultSecretIDKey)
}

func (vlt VaultConfiguration) TokenName() string {
	return viper.GetString(vaultTokenNameKey)
}
