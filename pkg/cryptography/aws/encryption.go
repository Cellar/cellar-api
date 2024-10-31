package aws

import (
	"cellar/pkg/models"
	"cellar/pkg/settings/cryptography"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	log "github.com/sirupsen/logrus"
)

type EncryptionClient struct {
	session       *session.Session
	kmsClient     *kms.KMS
	configuration cryptography.IAwsConfiguration
	logger        *log.Entry
}

func NewEncryptionClient(configuration cryptography.IAwsConfiguration) (*EncryptionClient, error) {
	logger := initializeLogger(configuration)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(configuration.Region()),
	})
	if err != nil {
		return nil, err
	}

	kmsClient := kms.New(sess)

	return &EncryptionClient{
		session:       sess,
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
	version := ec.kmsClient.APIVersion

	if version != "" {
		status = models.Healthy
	}

	return *models.NewHealth(name, status, version)
}

func (ec EncryptionClient) Encrypt(plaintext []byte) (ciphertext string, err error) {
	ec.logger.Debug("attempting to encrypt content")
	result, err := ec.kmsClient.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(ec.configuration.KmsKeyName()),
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
	result, err := ec.kmsClient.Decrypt(&kms.DecryptInput{
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
