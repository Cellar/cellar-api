package aws

import (
	pkgerrors "cellar/pkg/errors"
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

func NewEncryptionClient(ctx context.Context, configuration cryptography.IAwsConfiguration) (*EncryptionClient, error) {
	logger := initializeLogger(configuration)

	cfg, err := config.LoadDefaultConfig(ctx,
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

	logger.Debug("initializing aws kms configuration")
	if configuration.KmsKeyName() == "" {
		logger.Warn("AWS KMS key name is empty")
	}

	return logger
}

func (ec EncryptionClient) Health(ctx context.Context) models.Health {
	name := "AWS KMS"
	status := models.HealthStatus(models.Unhealthy)
	version := "2014-11-01" // KMS API version

	// Check context before operation
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return *models.NewHealth(name, status, version)
	}

	// Test connection with a simple operation
	keyId := ec.configuration.KmsKeyName()
	_, err := ec.kmsClient.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: &keyId,
	})

	if err == nil {
		status = models.Healthy
	}

	return *models.NewHealth(name, status, version)
}

func (ec EncryptionClient) Encrypt(ctx context.Context, plaintext []byte) (ciphertext string, err error) {
	// Check context before expensive operation
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return "", err
	}

	ec.logger.Debug("attempting to encrypt content")
	keyId := ec.configuration.KmsKeyName()
	result, err := ec.kmsClient.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     &keyId,
		Plaintext: plaintext,
	})

	if err != nil {
		ec.logger.WithError(err).
			Error("error encrypting content")

		return "", err
	}

	ciphertext = string(result.CiphertextBlob)

	return ciphertext, nil
}

func (ec EncryptionClient) Decrypt(ctx context.Context, ciphertext string) (plaintext []byte, err error) {
	// Check context before expensive operation
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil, err
	}

	ec.logger.Debug("attempting to decrypt content")
	result, err := ec.kmsClient.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: []byte(ciphertext),
	})

	if err != nil {
		ec.logger.WithError(err).
			Error("error decrypting content")
		return nil, err
	}

	plaintext = result.Plaintext

	return plaintext, nil
}
