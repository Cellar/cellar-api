package cryptography

import (
	"errors"
	"github.com/spf13/viper"
)

const (
	awsKey        = cryptographyKey + "aws."
	awsEnabledKey = awsKey + "enabled"

	awsRegionKey     = awsKey + "region"
	awsKmsKeyNameKey = awsKey + "kms_key_name"
)

type (
	AwsConfiguration  struct{}
	IAwsConfiguration interface {
		Region() string
		KmsKeyName() string
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

func (aws AwsConfiguration) KmsKeyName() string {
	return viper.GetString(awsKmsKeyNameKey)
}

func (aws AwsConfiguration) Validate() error {
	if aws.Region() == "" {
		return errors.New("AWS region not set")
	}
	if aws.KmsKeyName() == "" {
		return errors.New("AWS KMS key name not set")
	}

	return nil
}
