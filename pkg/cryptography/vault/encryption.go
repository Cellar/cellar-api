package vault

import (
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/models"
	"cellar/pkg/settings/cryptography"
	"context"
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
	configuration cryptography.IVaultConfiguration
	logger        *log.Entry
}

func NewEncryptionClient(ctx context.Context, configuration cryptography.IVaultConfiguration) (*EncryptionClient, error) {
	logger, err := initializeLogger(configuration)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	client, err := api.NewClient(&api.Config{
		Address:    configuration.Address(),
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

func initializeLogger(configuration cryptography.IVaultConfiguration) (*log.Entry, error) {
	logger := log.WithFields(log.Fields{
		"context":  "encryption",
		"instance": "vault",
		"host":     configuration.Address(),
	})

	logger.Debug("initializing vault configuration")
	if _, err := configuration.AuthConfiguration(); err != nil {
		logger.WithError(err).
			Error("vault auth configuration is invalid")
		return nil, err
	}
	if configuration.EncryptionTokenName() == "" {
		logger.Warn("vault token name is empty")
	}

	return logger, nil
}

func (vault EncryptionClient) Health(ctx context.Context) models.Health {
	name := "Vault"
	status := models.HealthStatus(models.Unhealthy)
	version := "Unknown"

	if err := pkgerrors.CheckContext(ctx); err != nil {
		return *models.NewHealth(name, status, version)
	}

	err := vault.login(ctx)
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

func (vault EncryptionClient) Encrypt(ctx context.Context, plaintext []byte) (ciphertext string, err error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return "", err
	}

	err = vault.login(ctx)
	if err != nil {
		return "", err
	}

	base64Content := base64.StdEncoding.EncodeToString(plaintext)

	vault.logger.Debug("attempting to encrypt content with vault")
	path := fmt.Sprintf("transit/encrypt/%s", vault.configuration.EncryptionTokenName())
	response, err := vault.client.Logical().Write(path, map[string]interface{}{
		"plaintext": base64Content,
	})

	if err != nil {
		vault.logger.WithError(err).
			Error("error encrypting content with vault")
		return "", err
	}

	if val, ok := response.Data["ciphertext"]; ok {
		vault.logger.Debug("content encryption successful")
		ciphertext = val.(string)
		return ciphertext, nil
	}

	return "", errors.New("unexpected response while encrypting secret")
}

func (vault EncryptionClient) Decrypt(ctx context.Context, ciphertext string) (plaintext []byte, err error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil, err
	}

	err = vault.login(ctx)
	if err != nil {
		return nil, err
	}

	vault.logger.Debug("attempting to decrypt content with vault")
	path := fmt.Sprintf("transit/decrypt/%s", vault.configuration.EncryptionTokenName())
	response, err := vault.client.Logical().Write(path, map[string]interface{}{
		"ciphertext": ciphertext,
	})

	if err != nil {
		vault.logger.WithError(err).
			Error("error decrypting content with vault")
		return nil, err
	}

	if val, ok := response.Data["plaintext"]; ok {
		base64Content := val.(string)
		if bytes, err := base64.StdEncoding.DecodeString(base64Content); err != nil {
			vault.logger.WithError(err).
				Error("error base64 decoding decrypted content")
		} else {
			vault.logger.Debug("content decryption successful")
			return bytes, nil
		}
	}

	return nil, errors.New("unexpected response while decrypting secret")
}
