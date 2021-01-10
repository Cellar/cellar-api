package settings

import (
	"cellar/pkg/aws"
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

const (
	vaultKey          = "vault."
	vaultTokenNameKey = vaultKey + "token_name"
	vaultAddressKey   = vaultKey + "address"
	vaultAuthBackend  = vaultKey + "auth_backend"

	vaultApprole            = vaultKey + "approle."
	vaultAppRoleRoleIdKey   = vaultApprole + "role_id"
	vaultAppRoleSecretIdKey = vaultApprole + "secret_id"

	vaultAwsIam     = vaultKey + "awsiam."
	vaultAwsIamRole = vaultAwsIam + "role"
)

type (
	VaultConfiguration  struct{}
	IVaultConfiguration interface {
		Address() string
		TokenName() string
		AuthBackend() (IAuthBackend, error)
	}
	IAuthBackend interface {
		LoginPath() string
		LoginParameters() map[string]interface{}
	}
	AwsIamAuthBackend struct {
		Role           string
		RequestMethod  string
		RequestUrl     string
		RequestBody    string
		RequestHeaders string
	}
	AppRoleAuthBackend struct {
		RoleId   string
		SecretId string
	}
)

func NewVaultConfiguration() *VaultConfiguration {
	viper.SetDefault(vaultAddressKey, "http://localhost:8200")
	return &VaultConfiguration{}
}

func NewAppRoleAuthBackend() (*AppRoleAuthBackend, error) {
	roleId := viper.GetString(vaultAppRoleRoleIdKey)
	secretId := viper.GetString(vaultAppRoleSecretIdKey)
	if secretId == "" && roleId == "" {
		return nil, nil
	} else if roleId == "" {
		return nil, errors.New("AppRole Role ID is empty")
	} else if secretId == "" {
		return nil, errors.New("AppRole Secret ID is empty")
	} else {
		return &AppRoleAuthBackend{
			RoleId:   roleId,
			SecretId: secretId,
		}, nil
	}
}

func NewAwsIamAuthBackend() (*AwsIamAuthBackend, error) {
	role := viper.GetString(vaultAwsIamRole)
	if role == "" {
		return nil, errors.New("AWS IAM Role is empty")
	}
	requestInfo, err := aws.GetAwsIamRequestInfo(viper.GetString(vaultAuthBackend))
	if err != nil {
		return nil, err
	}
	return &AwsIamAuthBackend{
		Role:           role,
		RequestMethod:  requestInfo.Method,
		RequestUrl:     requestInfo.Url,
		RequestBody:    requestInfo.Body,
		RequestHeaders: requestInfo.Headers,
	}, nil
}

func (vlt VaultConfiguration) Address() string {
	return viper.GetString(vaultAddressKey)
}

func (vlt VaultConfiguration) TokenName() string {
	return viper.GetString(vaultTokenNameKey)
}

func (vlt VaultConfiguration) AuthBackend() (IAuthBackend, error) {
	if backend, err := NewAppRoleAuthBackend(); backend != nil || err != nil {
		return backend, err
	}
	if backend, err := NewAwsIamAuthBackend(); backend != nil || err != nil {
		return backend, err
	}

	return nil, errors.New("no Vault auth backends were configured")
}

func (appRole AppRoleAuthBackend) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", viper.GetString(vaultAuthBackend))
}

func (appRole AppRoleAuthBackend) LoginParameters() map[string]interface{} {
	return map[string]interface{}{
		"role_id":   viper.GetString(vaultAppRoleRoleIdKey),
		"secret_id": viper.GetString(vaultAppRoleSecretIdKey),
	}
}

func (awsIam AwsIamAuthBackend) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", viper.GetString(vaultAuthBackend))
}

func (awsIam AwsIamAuthBackend) LoginParameters() map[string]interface{} {
	return map[string]interface{}{
		"role":                    awsIam.Role,
		"iam_http_request_method": awsIam.RequestMethod,
		"iam_request_url":         awsIam.RequestUrl,
		"iam_request_headers":     awsIam.RequestHeaders,
		"iam_request_body":        awsIam.RequestBody,
	}
}
