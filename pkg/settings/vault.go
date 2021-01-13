package settings

import (
	"cellar/pkg/aws"
	"cellar/pkg/gcp"
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

const (
	vaultKey          = "vault."
	vaultTokenNameKey = vaultKey + "token_name"
	vaultAddressKey   = vaultKey + "address"

	vaultAuth          = vaultKey + "auth."
	vaultAuthMountPath = vaultAuth + "mount_path"

	vaultApprole            = vaultAuth + "approle."
	vaultAppRoleRoleIdKey   = vaultApprole + "role_id"
	vaultAppRoleSecretIdKey = vaultApprole + "secret_id"

	vaultAwsIam     = vaultAuth + "awsiam."
	vaultAwsIamRole = vaultAwsIam + "role"

	vaultGcpIam     = vaultAuth + "gcpiam."
	vaultGcpIamRole = vaultGcpIam + "role"
)

type (
	VaultConfiguration  struct{}
	IVaultConfiguration interface {
		Address() string
		TokenName() string
		AuthConfiguration() (IVaultAuthConfiguration, error)
	}
	IVaultAuthConfiguration interface {
		Empty() bool
		Validate() error
		LoginPath() string
		LoginParameters() (map[string]interface{}, error)
	}
	AppRoleAuthBackend struct {
		RoleId   string
		SecretId string
	}
	AwsIamAuthBackend struct {
		Role string
	}
	GcpIamAuthBackend struct {
		Role string
	}
)

func NewVaultConfiguration() *VaultConfiguration {
	viper.SetDefault(vaultAddressKey, "http://localhost:8200")
	return &VaultConfiguration{}
}

func (vlt VaultConfiguration) Address() string {
	return viper.GetString(vaultAddressKey)
}

func (vlt VaultConfiguration) TokenName() string {
	return viper.GetString(vaultTokenNameKey)
}

func (vlt VaultConfiguration) AuthConfiguration() (IVaultAuthConfiguration, error) {
	var authBackends = []IVaultAuthConfiguration{
		NewAppRoleAuthBackend(),
		NewAwsIamAuthBackend(),
		NewGcpIamAuthBackend(),
	}
	var backend IVaultAuthConfiguration = nil
	for _, authBackend := range authBackends {
		if !authBackend.Empty() {
			if backend == nil {
				backend = authBackend
			} else {
				return nil, errors.New("only one vault auth method configuration is allowed but multiple were detected")
			}
		}
	}
	if backend == nil {
		return nil, errors.New("no Vault auth methods configurations were detected")
	}

	if err := backend.Validate(); err != nil {
		return nil, err
	}

	return backend, nil
}

/**********
* APPROLE *
**********/
func NewAppRoleAuthBackend() *AppRoleAuthBackend {
	return &AppRoleAuthBackend{
		RoleId:   viper.GetString(vaultAppRoleRoleIdKey),
		SecretId: viper.GetString(vaultAppRoleSecretIdKey),
	}
}

func (appRole AppRoleAuthBackend) Empty() bool {
	return appRole.RoleId == "" && appRole.SecretId == ""
}

func (appRole AppRoleAuthBackend) Validate() error {
	if appRole.RoleId == "" {
		return errors.New("AppRole Role ID is empty")
	}
	if appRole.SecretId == "" {
		return errors.New("AppRole Secret ID is empty")
	}
	return nil
}

func (appRole AppRoleAuthBackend) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", viper.GetString(vaultAuthMountPath))
}

func (appRole AppRoleAuthBackend) LoginParameters() (map[string]interface{}, error) {
	return map[string]interface{}{
		"role_id":   viper.GetString(vaultAppRoleRoleIdKey),
		"secret_id": viper.GetString(vaultAppRoleSecretIdKey),
	}, nil
}

/**********
* AWS IAM *
**********/
func NewAwsIamAuthBackend() *AwsIamAuthBackend {
	return &AwsIamAuthBackend{
		Role: viper.GetString(vaultAwsIamRole),
	}
}

func (awsIam AwsIamAuthBackend) Empty() bool {
	return awsIam.Role == ""
}

func (awsIam AwsIamAuthBackend) Validate() error {
	if awsIam.Role == "" {
		return errors.New("AWS IAM Role is empty")
	}
	return nil
}

func (awsIam AwsIamAuthBackend) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", viper.GetString(vaultAuthMountPath))
}

func (awsIam AwsIamAuthBackend) LoginParameters() (map[string]interface{}, error) {
	requestInfo, err := aws.GetAwsIamRequestInfo(viper.GetString(vaultAuthMountPath))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"role":                    awsIam.Role,
		"iam_http_request_method": requestInfo.Method,
		"iam_request_url":         requestInfo.Url,
		"iam_request_headers":     requestInfo.Headers,
		"iam_request_body":        requestInfo.Body,
	}, nil
}

/**********
* GCP IAM *
**********/
func NewGcpIamAuthBackend() *GcpIamAuthBackend {
	return &GcpIamAuthBackend{
		Role: viper.GetString(vaultGcpIamRole),
	}
}

func (gcpIam GcpIamAuthBackend) Empty() bool {
	return gcpIam.Role == ""
}

func (gcpIam GcpIamAuthBackend) Validate() error {
	if gcpIam.Role == "" {
		return errors.New("AWS IAM Role is empty")
	}
	return nil
}

func (gcpIam GcpIamAuthBackend) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", viper.GetString(vaultAuthMountPath))
}

func (gcpIam GcpIamAuthBackend) LoginParameters() (map[string]interface{}, error) {
	signedJwt, err := gcp.GetGcpIamRequestInfo(gcpIam.Role)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"role": gcpIam.Role,
		"jwt":  signedJwt,
	}, nil
}
