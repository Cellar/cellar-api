package commands

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/models"
	"crypto/rand"
	"encoding/hex"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

func getLogger(secretId string) *log.Entry {
	return log.WithFields(log.Fields{
		"context":  "secret commands",
		"secretId": secretId,
	})
}

func CreateSecret(dataStore datastore.DataStore, encryption cryptography.Encryption, secret models.Secret) (response *models.SecretMetadata, isValidationError bool, err error) {
	isValidationError = false

	id, err := randomId()
	if err != nil {
		return
	}

	logger := getLogger(id)
	logger.Info("Encrypting new secret content")

	secret.CipherText, err = encryption.Encrypt(secret.Content)
	if err != nil {
		logger.WithError(err).
			Error("Error encrypting new secret content")
		return
	}

	secret.ID = id
	secret.AccessCount = 0
	if secret.AccessLimit < 0 { secret.AccessLimit = 0 }

	if secret.Duration() < time.Minute*10 {
		return response, true, errors.New("expiration must be at least 10 minutes in the future")
	}

	logger = logger.WithFields(log.Fields{
		"secretAccessLimit": secret.AccessLimit,
		"secretExpiration":  secret.Expiration().Format(),
	})
	logger.Info("Writing new secret to datastore")
	err = dataStore.WriteSecret(secret)
	if err != nil {
		logger.WithError(err).Error("Error writing new secret to datastore")
		return
	}

	response = secret.Metadata()
	return
}

func AccessSecret(dataStore datastore.DataStore, encryption cryptography.Encryption, id string) (*models.Secret, error) {

	secret := dataStore.ReadSecret(id)
	if secret == nil {
		return nil, nil
	}

	accessCount, err := dataStore.IncreaseAccessCount(id)
	if err != nil {
		return nil, err
	}

	logger := getLogger(id).
		WithFields(log.Fields{
			"secretAccessCount": accessCount,
			"secretAccessLimit": secret.AccessLimit,
			"secretExpiration":  secret.Expiration().Format(),
		})
	logger.Info("Accessed secret")

	if secret.AccessLimit > 0 && accessCount >= int64(secret.AccessLimit) {
		logger.Info("Deleting secret with access limit reached")
		if _, err = dataStore.DeleteSecret(id); err != nil {
			logger.WithError(err).
				Error("Error while deleting secret")
			return nil, err
		}
	}

	content, err := encryption.Decrypt(secret.CipherText)
	if err != nil {
		return nil, err
	}

	return &models.Secret{
		ID:      id,
		Content: content,
		ContentType: secret.ContentType,
	}, nil
}

func GetSecretMetadata(dataStore datastore.DataStore, id string) *models.SecretMetadata {
	logger := getLogger(id)
	logger.Info("Querying for secret metadata")

	secret := dataStore.ReadSecret(id)
	if secret == nil {
		return nil
	}

	return secret.Metadata()
}

func DeleteSecret(dataStore datastore.DataStore, id string) (bool, error) {
	getLogger(id).Info("Deleting secret if it exists")
	return dataStore.DeleteSecret(id)
}

func randomId() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		log.WithField("context", "secretcommand").
			WithError(err).
			Error("Error while generating new ID for secret")
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
