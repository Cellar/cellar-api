package commands

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/models"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
)

func getLogger(secretId string) *log.Entry {
	return log.WithFields(log.Fields{
		"context":  "secret commands",
		"secretId": secretId,
	})
}

// CreateSecret encrypts and stores a new secret with the given parameters.
// Returns the secret metadata, a validation error flag, and any error encountered.
// The context can be used to cancel the operation before completion.
func CreateSecret(ctx context.Context, dataStore datastore.DataStore, encryption cryptography.Encryption, secret models.Secret) (response *models.SecretMetadata, isValidationError bool, err error) {
	if err = pkgerrors.CheckContext(ctx); err != nil {
		return
	}

	isValidationError = false

	id, err := randomId()
	if err != nil {
		return
	}

	logger := getLogger(id)
	logger.Info("Encrypting new secret content")

	secret.CipherText, err = encryption.Encrypt(ctx, secret.Content)
	if err != nil {
		logger.WithError(err).
			Error("Error encrypting new secret content")
		return
	}

	secret.ID = id
	secret.AccessCount = 0
	if secret.AccessLimit < 0 {
		secret.AccessLimit = 0
	}

	if secret.Duration() < time.Minute*10 {
		return response, true, errors.New("expiration must be at least 10 minutes in the future")
	}

	logger = logger.WithFields(log.Fields{
		"secretAccessLimit": secret.AccessLimit,
		"secretExpiration":  secret.Expiration().Format(),
	})
	logger.Info("Writing new secret to datastore")
	err = dataStore.WriteSecret(ctx, secret)
	if err != nil {
		logger.WithError(err).Error("Error writing new secret to datastore")
		return
	}

	response = secret.Metadata()
	return
}

// AccessSecret retrieves and decrypts a secret by ID, incrementing its access count.
// If the access limit is reached, the secret is automatically deleted.
// Returns the decrypted secret or nil if not found.
// The context can be used to cancel the operation before completion.
func AccessSecret(ctx context.Context, dataStore datastore.DataStore, encryption cryptography.Encryption, id string) (*models.Secret, error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil, err
	}

	secret := dataStore.ReadSecret(ctx, id)
	if secret == nil {
		return nil, nil
	}

	accessCount, err := dataStore.IncreaseAccessCount(ctx, id)
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
		if _, err = dataStore.DeleteSecret(ctx, id); err != nil {
			logger.WithError(err).
				Error("Error while deleting secret")
			return nil, err
		}
	}

	content, err := encryption.Decrypt(ctx, secret.CipherText)
	if err != nil {
		return nil, err
	}

	return &models.Secret{
		ID:          id,
		Content:     content,
		ContentType: secret.ContentType,
	}, nil
}

// GetSecretMetadata retrieves metadata for a secret without decrypting its content.
// Returns the metadata or nil if the secret is not found.
// The context can be used to cancel the operation before completion.
func GetSecretMetadata(ctx context.Context, dataStore datastore.DataStore, id string) *models.SecretMetadata {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil
	}

	logger := getLogger(id)
	logger.Info("Querying for secret metadata")

	secret := dataStore.ReadSecret(ctx, id)
	if secret == nil {
		return nil
	}

	return secret.Metadata()
}

// DeleteSecret removes a secret from the datastore by ID.
// Returns true if the secret was found and deleted, false if not found.
// The context can be used to cancel the operation before completion.
func DeleteSecret(ctx context.Context, dataStore datastore.DataStore, id string) (bool, error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return false, err
	}

	getLogger(id).Info("Deleting secret if it exists")
	return dataStore.DeleteSecret(ctx, id)
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
