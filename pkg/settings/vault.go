package settings

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

const (
	vaultKey                = "vault."
	vaultTokenNameKey       = vaultKey + "token_name"
	vaultAddressKey         = vaultKey + "address"
	vaultAuthBackend        = vaultKey + "auth_backend"
	vaultApprole            = vaultKey + "approle."
	vaultAppRoleRoleIDKey   = vaultApprole + "role_id"
	vaultAppRoleSecretIDKey = vaultApprole + "secret_id"
)

type (
	VaultConfiguration struct{}
	IVaultConfiguration interface {
		Address() string
		TokenName() string
		AuthBackend() (IAuthBackend, error)
	}
	IAuthBackend interface {
		LoginPath() string
		LoginParameters() map[string]interface{}
	}
	AppRoleAuthBackend struct {
		RoleId string
		SecretId string
	}
)

func NewVaultConfiguration() *VaultConfiguration {
	viper.SetDefault(vaultAddressKey, "http://localhost:8200")
	return &VaultConfiguration{}
}

func NewAppRoleAuthBackend() (*AppRoleAuthBackend, error) {
	roleId := viper.GetString(vaultAppRoleRoleIDKey)
	secretId := viper.GetString(vaultAppRoleSecretIDKey)
	if roleId == "" {
		return nil, errors.New("AppRole Role ID is empty")
	}
	if secretId == "" {
		return nil, errors.New("AppRole Secret ID is empty")
	}
	return &AppRoleAuthBackend{
		RoleId:   roleId,
		SecretId: secretId,
	}, nil
}

func (vlt VaultConfiguration) Address() string {
	return viper.GetString(vaultAddressKey)
}

func (vlt VaultConfiguration) TokenName() string {
	return viper.GetString(vaultTokenNameKey)
}

func (vlt VaultConfiguration) AuthBackend() (authBackend IAuthBackend, err error) {
	return NewAppRoleAuthBackend()
}

func (appRole AppRoleAuthBackend) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", viper.GetString(vaultAuthBackend))
}

func (appRole AppRoleAuthBackend) LoginParameters() map[string]interface{} {
	return map[string]interface{}{
		"role_id": viper.GetString(vaultAppRoleRoleIDKey),
		"secret_id": viper.GetString(vaultAppRoleSecretIDKey),
	}
}
