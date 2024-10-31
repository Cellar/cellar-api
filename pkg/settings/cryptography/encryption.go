package cryptography

type IEncryptionConfiguration interface {
	Vault() IVaultConfiguration
	Aws() IAwsConfiguration
}

const (
	cryptographyKey = "cryptography."
)

type EncryptionConfiguration struct{}

func NewEncryptionConfiguration() *EncryptionConfiguration {
	return &EncryptionConfiguration{}
}

func (e *EncryptionConfiguration) Vault() IVaultConfiguration {
	return NewVaultConfiguration()
}

func (e *EncryptionConfiguration) Aws() IAwsConfiguration {
	return NewAwsConfiguration()
}
