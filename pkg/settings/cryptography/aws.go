package cryptography

import (
	"errors"

	"github.com/spf13/viper"
)

const (
	awsKey        = cryptographyKey + "aws."
	awsEnabledKey = awsKey + "enabled"

	awsRegionKey   = awsKey + "region"
	awsKmsKeyIdKey = awsKey + "kms_key_id"
)

type (
	AwsConfiguration  struct{}
	IAwsConfiguration interface {
		Region() string
		KmsKeyId() string
		Enabled() bool
		Validate() error
	}
)

func NewAwsConfiguration() *AwsConfiguration {
	return &AwsConfiguration{}
}

func (aws AwsConfiguration) Enabled() bool {
	return viper.GetBool(awsEnabledKey)
}

func (aws AwsConfiguration) Region() string {
	return viper.GetString(awsRegionKey)
}

func (aws AwsConfiguration) KmsKeyId() string {
	return viper.GetString(awsKmsKeyIdKey)
}

func (aws AwsConfiguration) Validate() error {
	if aws.Region() == "" {
		return errors.New("AWS region not set")
	}
	if aws.KmsKeyId() == "" {
		return errors.New("AWS KMS key ID not set")
	}

	return nil
}
