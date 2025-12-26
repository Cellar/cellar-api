package aws

import (
	"cellar/pkg/models"
	"cellar/pkg/settings/cryptography"
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	log "github.com/sirupsen/logrus"
)

type EncryptionClient struct {
	kmsClient     *kms.Client
	configuration cryptography.IAwsConfiguration
	logger        *log.Entry
}

func NewEncryptionClient(configuration cryptography.IAwsConfiguration) (*EncryptionClient, error) {
	logger := initializeLogger(configuration)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(configuration.Region()),
	)
	if err != nil {
		return nil, err
	}

	kmsClient := kms.NewFromConfig(cfg)

	return &EncryptionClient{
		kmsClient:     kmsClient,
		configuration: configuration,
		logger:        logger,
	}, nil
}

func initializeLogger(configuration cryptography.IAwsConfiguration) *log.Entry {
	logger := log.WithFields(log.Fields{
		"context":  "encryption",
		"instance": "aws",
	})

	logger.Debug("initializing vault configuration")
	if configuration.KmsKeyName() == "" {
		logger.Warn("AWS KMS key name is empty")
	}

	return logger
}

func (ec EncryptionClient) Health() models.Health {
	name := "AWS KMS"
	status := models.HealthStatus(models.Unhealthy)
	version := "2014-11-01" // KMS API version

	// Test connection with a simple operation
	keyId := ec.configuration.KmsKeyName()
	_, err := ec.kmsClient.DescribeKey(context.TODO(), &kms.DescribeKeyInput{
		KeyId: &keyId,
	})

	if err == nil {
		status = models.Healthy
	}

	return *models.NewHealth(name, status, version)
}

func (ec EncryptionClient) Encrypt(plaintext []byte) (ciphertext string, err error) {
	ec.logger.Debug("attempting to encrypt content")
	keyId := ec.configuration.KmsKeyName()
	result, err := ec.kmsClient.Encrypt(context.TODO(), &kms.EncryptInput{
		KeyId:     &keyId,
		Plaintext: plaintext,
	})

	if err != nil {
		ec.logger.WithError(err).
			Error("error encrypting content")

		return
	}

	ciphertext = string(result.CiphertextBlob)

	return
}

func (ec EncryptionClient) Decrypt(ciphertext string) (plaintext []byte, err error) {
	ec.logger.Debug("attempting to decrypt content")
	result, err := ec.kmsClient.Decrypt(context.TODO(), &kms.DecryptInput{
		CiphertextBlob: []byte(ciphertext),
	})

	if err != nil {
		ec.logger.WithError(err).
			Error("error decrypting content")
		return
	}

	plaintext = result.Plaintext

	return
}
