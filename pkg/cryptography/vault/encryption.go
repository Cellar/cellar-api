package vault

import (
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type EncryptionClient struct {
	client        *api.Client
	configuration settings.IConfiguration
	logger        *log.Entry
}

func NewEncryptionClient(configuration settings.IConfiguration) (*EncryptionClient, error) {
	logger, err := initializeLogger(configuration)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	client, err := api.NewClient(&api.Config{
		Address:    configuration.Vault().Address(),
		HttpClient: httpClient,
	})
	if err != nil {
		return nil, err
	}

	return &EncryptionClient{
		client:        client,
		configuration: configuration,
		logger:        logger,
	}, nil
}

func initializeLogger(configuration settings.IConfiguration) (*log.Entry, error) {
	logger := log.WithFields(log.Fields{
		"context":  "encryption",
		"instance": "vault",
		"host":     configuration.Vault().Address(),
	})

	logger.Debug("initializing vault configuration")
	if _, err := configuration.Vault().AuthBackend(); err != nil {
		logger.WithError(err).
			Error("vault auth configuration is invalid")
		return nil, err
	}
	if configuration.Vault().TokenName() == "" {
		logger.Warn("vault token name is empty")
	}

	return logger, nil
}

func (vault EncryptionClient) Health() models.Health {
	name := "Vault"
	status := models.HealthStatus(models.Unhealthy)
	version := "Unknown"

	err := vault.login()
	if err == nil {
		res, err := vault.client.Sys().Health()
		if err == nil {
			version = res.Version
			if res.Sealed {
				status = models.Degraded
			} else {
				status = models.Healthy
			}
		}
	}

	return *models.NewHealth(name, status, version)
}

func (vault EncryptionClient) Encrypt(content string) (encryptedContent string, err error) {
	err = vault.login()
	if err != nil {
		return
	}

	base64Content := base64.StdEncoding.EncodeToString([]byte(content))

	vault.logger.Debug("attempting to encrypt content with vault")
	path := fmt.Sprintf("transit/encrypt/%s", vault.configuration.Vault().TokenName())
	response, err := vault.client.Logical().Write(path, map[string]interface{}{
		"plaintext": base64Content,
	})

	if err != nil {
		vault.logger.WithError(err).
			Error("error encrypting content with vault")
		return
	}

	if val, ok := response.Data["ciphertext"]; ok {
		vault.logger.Debug("content encryption successful")
		encryptedContent = val.(string)
		return
	}

	return "", errors.New("unexpected response while encrypting secret")
}

func (vault EncryptionClient) Decrypt(content string) (decryptedContent string, err error) {
	err = vault.login()
	if err != nil {
		return
	}

	vault.logger.Debug("attempting to decrypt content with vault")
	path := fmt.Sprintf("transit/decrypt/%s", vault.configuration.Vault().TokenName())
	response, err := vault.client.Logical().Write(path, map[string]interface{}{
		"ciphertext": content,
	})

	if err != nil {
		vault.logger.WithError(err).
			Error("error decrypting content with vault")
		return
	}

	if val, ok := response.Data["plaintext"]; ok {
		base64Content := val.(string)
		bytes, err := base64.StdEncoding.DecodeString(base64Content)
		vault.logger.Debug("content decryption successful")
		return string(bytes), err
	}

	return "", errors.New("unexpected response while decrypting secret")
}
