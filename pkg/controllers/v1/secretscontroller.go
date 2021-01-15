package v1

import (
	"cellar/pkg/commands"
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
	"net/http"
)

// @Summary Create Secret
// @Produce json
// @Accept json
// @Param secret body models.CreateSecretRequest true "Add secret"
// @Success 201 {object} models.SecretMetadataResponse
// @Failure 400 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /v1/secrets [post]
func CreateSecret(c *gin.Context) {
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)
	encryption := c.MustGet(cryptography.Key).(cryptography.Encryption)

	var body models.CreateSecretRequest
	var secret models.Secret
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	if body.Content == nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("required parameter: content"))
		return
	} else {
		secret.Content = []byte(*body.Content)
		secret.ContentType = string(models.ContentTypeText)
	}

	if body.ExpirationEpoch == nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("required parameter: duration"))
		return
	} else {
		secret.ExpirationEpoch = *body.ExpirationEpoch
	}

	if body.AccessLimit == nil {
		secret.AccessLimit = 0
	} else {
		secret.AccessLimit = *body.AccessLimit
	}

	if response, isValidationError, err := commands.CreateSecretV2(dataStore, encryption,secret); err != nil {
		if isValidationError {
			httputil.NewError(c, http.StatusBadRequest, err)
		} else {
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	} else {
		c.JSON(http.StatusCreated, response)
	}
}

// @Summary Access Secret Content
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

	if secret, err := commands.AccessSecret(dataStore, encryption, id); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	} else if secret == nil {
		c.Status(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, models.SecretContentResponse{
			ID:      secret.ID,
			Content: secret.Content,
		})
	}
}

// @Summary Get Secret Metadata
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

	if secretMetadata := commands.GetSecretMetadata(dataStore, id); secretMetadata == nil {
		c.Status(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, secretMetadata)
	}
}

// @Summary Delete Secret
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

	if deleted, err := commands.DeleteSecret(dataStore, id); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	} else if !deleted {
		c.Status(http.StatusNotFound)
	} else {
		c.Status(http.StatusNoContent)
	}
}
