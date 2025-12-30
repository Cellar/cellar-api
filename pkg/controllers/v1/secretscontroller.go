package v1

import (
	"cellar/pkg/commands"
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Create Secret
// @Tags v1
// @Produce json
// @Accept json
// @Param secret body models.CreateSecretRequest true "Add secret"
// @Success 201 {object} models.SecretMetadataResponse
// @Failure 400 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /v1/secrets [post]
func CreateSecret(c *gin.Context) {
	cfg := c.MustGet(settings.Key).(settings.IConfiguration)
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)
	encryption := c.MustGet(cryptography.Key).(cryptography.Encryption)

	var body models.CreateSecretRequest
	var secret models.Secret
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.Error(pkgerrors.NewValidationError(err.Error()))
		return
	}

	if body.Content == nil {
		_ = c.Error(pkgerrors.NewValidationError("required parameter: content"))
		return
	}
	secret.Content = []byte(*body.Content)
	secret.ContentType = string(models.ContentTypeText)

	if body.ExpirationEpoch == nil {
		_ = c.Error(pkgerrors.NewValidationError("required parameter: duration"))
		return
	}
	secret.ExpirationEpoch = *body.ExpirationEpoch

	if body.AccessLimit == nil {
		secret.AccessLimit = 0
	} else {
		secret.AccessLimit = *body.AccessLimit
	}

	metadata, err := commands.CreateSecret(context.Background(), cfg.App(), dataStore, encryption, secret)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, models.SecretMetadataResponse{
		ID:          metadata.ID,
		AccessCount: metadata.AccessCount,
		AccessLimit: metadata.AccessLimit,
		Expiration:  metadata.Expiration,
	})
}

// @Summary Access Secret Content
// @Tags v1
// @Produce json
// @Accept json
// @Param id path string true "Secret ID"
// @Success 200 {object} models.SecretContentResponse
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /v1/secrets/{id}/access [post]
func AccessSecretContent(c *gin.Context) {
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)
	encryption := c.MustGet(cryptography.Key).(cryptography.Encryption)

	id := c.Param("id")

	secret, err := commands.AccessSecret(context.Background(), dataStore, encryption, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if secret == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, models.SecretContentResponse{
		ID:      secret.ID,
		Content: string(secret.Content),
	})
}

// @Summary Get Secret Metadata
// @Tags v1
// @Produce json
// @Accept json
// @Param id path string true "Secret ID"
// @Success 200 {object} models.SecretMetadataResponse
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /v1/secrets/{id} [get]
func GetSecretMetadata(c *gin.Context) {
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)

	id := c.Param("id")

	if secretMetadata := commands.GetSecretMetadata(context.Background(), dataStore, id); secretMetadata == nil {
		c.Status(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, models.SecretMetadataResponse{
			ID:          secretMetadata.ID,
			AccessCount: secretMetadata.AccessCount,
			AccessLimit: secretMetadata.AccessLimit,
			Expiration:  secretMetadata.Expiration,
		})
	}
}

// @Summary Delete Secret
// @Tags v1
// @Produce json
// @Accept json
// @Param id path string true "Secret ID"
// @Success 204 ""
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /v1/secrets/{id} [delete]
func DeleteSecret(c *gin.Context) {
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)

	id := c.Param("id")

	deleted, err := commands.DeleteSecret(context.Background(), dataStore, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if !deleted {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}
