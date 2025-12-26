package cryptography

import (
	"cellar/pkg/aws"
	"cellar/pkg/gcp"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
)

const (
	vaultKey        = cryptographyKey + "vault."
	vaultEnabledKey = vaultKey + "enabled"

	vaultAddressKey          = vaultKey + "address"
	vaultEncryptionTokenName = vaultKey + "encryption_token_name"

	vaultAuth          = vaultKey + "auth."
	vaultAuthMountPath = vaultAuth + "mount_path"

	vaultApprole            = vaultAuth + "approle."
	vaultAppRoleRoleIdKey   = vaultApprole + "role_id"
	vaultAppRoleSecretIdKey = vaultApprole + "secret_id"

	vaultAwsIam     = vaultAuth + "awsiam."
	vaultAwsIamRole = vaultAwsIam + "role"

	vaultGcpIam     = vaultAuth + "gcpiam."
	vaultGcpIamRole = vaultGcpIam + "role"

	vaultKubernetes     = vaultAuth + "kubernetes."
	vaultKubernetesRole = vaultKubernetes + "role"
)

type (
	VaultConfiguration  struct{}
	IVaultConfiguration interface {
		Enabled() bool
		Address() string
		EncryptionTokenName() string
		AuthConfiguration() (IVaultAuthConfiguration, error)
	}
	IVaultAuthConfiguration interface {
		Empty() bool
		Validate() error
		LoginPath() string
		LoginParameters() (map[string]interface{}, error)
	}
	AppRoleAuth struct {
		MountPath string
		RoleId    string
		SecretId  string
	}
	AwsIamAuth struct {
		MountPath string
		Role      string
	}
	GcpIamAuth struct {
		MountPath string
		Role      string
	}
	KubernetesAuth struct {
		MountPath string
		Role      string
	}
)

func NewVaultConfiguration() *VaultConfiguration {
	viper.SetDefault(vaultAddressKey, "http://localhost:8200")
	return &VaultConfiguration{}
}

func (vlt VaultConfiguration) Enabled() bool {
	return viper.GetBool(vaultEnabledKey)
}

func (vlt VaultConfiguration) Validate() error {
	if vlt.Address() == "" {
		return errors.New("vault address not set")
	}
	if vlt.EncryptionTokenName() == "" {
		return errors.New("vault encryption token not set")
	}

	if _, err := vlt.AuthConfiguration(); err != nil {
		return err
	}

	return nil
}

func (vlt VaultConfiguration) Address() string {
	return viper.GetString(vaultAddressKey)
}

func (vlt VaultConfiguration) EncryptionTokenName() string {
	return viper.GetString(vaultEncryptionTokenName)
}

func (vlt VaultConfiguration) AuthConfiguration() (IVaultAuthConfiguration, error) {
	mountPath := viper.GetString(vaultAuthMountPath)
	if mountPath == "" {
		return nil, fmt.Errorf("%s is empty", vaultAuthMountPath)
	}
	var authBackends = []IVaultAuthConfiguration{
		NewAppRoleAuth(mountPath),
		NewAwsIamAuth(mountPath),
		NewGcpIamAuth(mountPath),
		NewKubernetesAuth(mountPath),
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
		return nil, errors.New("no Vault auth method configurations were detected")
	}

	if err := backend.Validate(); err != nil {
		return nil, err
	}

	return backend, nil
}

/*************
* APPROLE    *
*************/
func NewAppRoleAuth(mountPath string) *AppRoleAuth {
	return &AppRoleAuth{
		MountPath: mountPath,
		RoleId:    viper.GetString(vaultAppRoleRoleIdKey),
		SecretId:  viper.GetString(vaultAppRoleSecretIdKey),
	}
}

func (appRole AppRoleAuth) Empty() bool {
	return appRole.RoleId == "" && appRole.SecretId == ""
}

func (appRole AppRoleAuth) Validate() error {
	if appRole.RoleId == "" {
		return errors.New("AppRole Role ID is empty")
	}
	if appRole.SecretId == "" {
		return errors.New("AppRole Secret ID is empty")
	}
	return nil
}

func (appRole AppRoleAuth) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", appRole.MountPath)
}

func (appRole AppRoleAuth) LoginParameters() (map[string]interface{}, error) {
	return map[string]interface{}{
		"role_id":   viper.GetString(vaultAppRoleRoleIdKey),
		"secret_id": viper.GetString(vaultAppRoleSecretIdKey),
	}, nil
}

/*************
* AWS IAM    *
*************/
func NewAwsIamAuth(mountPath string) *AwsIamAuth {
	return &AwsIamAuth{
		MountPath: mountPath,
		Role:      viper.GetString(vaultAwsIamRole),
	}
}

func (awsIam AwsIamAuth) Empty() bool {
	return awsIam.Role == ""
}

func (awsIam AwsIamAuth) Validate() error {
	if awsIam.Role == "" {
		return errors.New("AWS IAM role is empty")
	}
	return nil
}

func (awsIam AwsIamAuth) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", awsIam.MountPath)
}

func (awsIam AwsIamAuth) LoginParameters() (map[string]interface{}, error) {
	requestInfo, err := aws.GetAwsIamRequestInfo(context.Background(), viper.GetString(vaultAuthMountPath))
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

/*************
* GCP IAM    *
*************/
func NewGcpIamAuth(mountPath string) *GcpIamAuth {
	return &GcpIamAuth{
		MountPath: mountPath,
		Role:      viper.GetString(vaultGcpIamRole),
	}
}

func (gcpIam GcpIamAuth) Empty() bool {
	return gcpIam.Role == ""
}

func (gcpIam GcpIamAuth) Validate() error {
	if gcpIam.Role == "" {
		return errors.New("GCP IAM role is empty")
	}
	return nil
}

func (gcpIam GcpIamAuth) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", gcpIam.MountPath)
}

func (gcpIam GcpIamAuth) LoginParameters() (map[string]interface{}, error) {
	signedJwt, err := gcp.GetGcpIamRequestInfo(gcpIam.Role)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"role": gcpIam.Role,
		"jwt":  signedJwt,
	}, nil
}

/*************
* Kubernetes *
*************/
func NewKubernetesAuth(mountPath string) *KubernetesAuth {
	return &KubernetesAuth{
		MountPath: mountPath,
		Role:      viper.GetString(vaultKubernetesRole),
	}
}

func (k8s KubernetesAuth) Empty() bool {
	return k8s.Role == ""
}

func (k8s KubernetesAuth) Validate() error {
	if k8s.Role == "" {
		return errors.New("kubernetes role is empty")
	}
	return nil
}

func (k8s KubernetesAuth) LoginPath() string {
	return fmt.Sprintf("auth/%s/login", k8s.MountPath)
}

func (k8s KubernetesAuth) LoginParameters() (map[string]interface{}, error) {
	jwtBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, err
	}
	jwt := string(jwtBytes)
	if jwt == "" {
		return nil, errors.New("kubernetes service account jwt was found to be empty")
	}
	return map[string]interface{}{
		"role": k8s.Role,
		"jwt":  jwt,
	}, nil
}
