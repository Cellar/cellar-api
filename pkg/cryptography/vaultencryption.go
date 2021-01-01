package cryptography

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

type VaultEncryption struct {
	client        *api.Client
	configuration settings.IConfiguration
	logger        *log.Entry
}

func NewVaultEncryption(configuration settings.IConfiguration) (*VaultEncryption, error) {
	logger := initializeLogger(configuration)

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

	return &VaultEncryption{
		client:        client,
		configuration: configuration,
		logger:        logger,
	}, nil
}

func initializeLogger(configuration settings.IConfiguration) *log.Entry {
	logger := log.WithFields(log.Fields{
		"context":  "encryption",
		"instance": "vault",
		"host":     configuration.Vault().Address(),
	})

	logger.Debug("initializing vault configuration")
	if configuration.Vault().RoleID() == "" {
		logger.Warn("vault role_id is empty")
	}
	if configuration.Vault().SecretID() == "" {
		logger.Warn("vault secret_id is empty")
	}
	if configuration.Vault().TokenName() == "" {
		logger.Warn("vault token name is empty")
	}

	return logger
}

func (vault VaultEncryption) Health() models.Health {
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

func (vault VaultEncryption) login() error {
	vault.logger.Debug("attempting to find and renew existing tokens")
	token, err := vault.client.Auth().Token().RenewSelf(60)
	if err == nil && token != nil {
		vault.logger.Debug("token renewal successful")
		vault.client.SetToken(token.Auth.ClientToken)
		return nil
	} else {
		vault.logger.Debug("unable to find or renew existing tokens")
	}

	vault.logger.Debug("attempting to login to vault")
	secret, err := vault.client.Logical().Write("auth/approle/login", map[string]interface{}{
		"role_id":   vault.configuration.Vault().RoleID(),
		"secret_id": vault.configuration.Vault().SecretID(),
	})
	if err != nil {
		vault.logger.WithError(err).
			Error("unable to login to vault")
		return err
	}

	vault.logger.Debug("login to vault successful")
	vault.client.SetToken(secret.Auth.ClientToken)
	return nil
}

func (vault VaultEncryption) Encrypt(content string) (encryptedContent string, err error) {
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

func (vault VaultEncryption) Decrypt(content string) (decryptedContent string, err error) {
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
