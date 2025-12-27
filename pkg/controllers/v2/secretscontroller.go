package v2

import (
	"bytes"
	"cellar/pkg/commands"
	"cellar/pkg/controllers"
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"cellar/pkg/validators"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

// @Summary Create Secret
// @Tags v2
// @Produce application/json
// @Accept multipart/form-data
// @Param content formData string false "Secret content"
// @Param access_limit formData int false "Access limit"
// @Param expiration_epoch formData int true "Expiration of the secret in Unix Epoch Time"
// @Param file formData file false "Secret content as a file"
// @Success 201 {object} models.SecretMetadataResponseV2
// @Failure 400 {object} httputil.HTTPError "Bad Request - validation error"
// @Failure 408 {object} httputil.HTTPError "Request Timeout - operation cancelled"
// @Failure 413 {object} httputil.HTTPError "Payload Too Large - file exceeds size limit"
// @Failure 500 {object} httputil.HTTPError
// @Router /v2/secrets [post]
func CreateSecret(c *gin.Context) {
	ctx := c.Request.Context()
	cfg := c.MustGet(settings.Key).(settings.IConfiguration)
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)
	encryption := c.MustGet(cryptography.Key).(cryptography.Encryption)

	var secret models.Secret

	if accessLimitStr := c.PostForm("access_limit"); accessLimitStr != "" {
		if accessLimit, err := strconv.Atoi(accessLimitStr); err != nil {
			httputil.NewError(c, http.StatusBadRequest, errors.New("optional parameter: access_limit: invalid value"))
			return
		} else {
			secret.AccessLimit = accessLimit
		}
	}

	if expirationEpoch, err := strconv.ParseInt(c.PostForm("expiration_epoch"), 10, 64); err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("required parameter: expiration_epoch"))
		return
	} else {
		secret.ExpirationEpoch = expirationEpoch
	}

	content := c.PostForm("content")
	fileHeader, err := c.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		httputil.NewError(c, http.StatusBadRequest, errors.New("required parameter: file: invalid value"))
		return
	}

	if content != "" && fileHeader != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("secret with both content and file is not allowed"))
		return
	} else if content == "" && fileHeader == nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("required parameter: file or content"))
		return
	} else if content != "" {
		secret.Content = []byte(content)
		secret.ContentType = models.ContentTypeText
	} else {
		if fileHeader.Size == 0 {
			httputil.NewError(c, http.StatusBadRequest, errors.New("file cannot be empty"))
			return
		}

		maxSizeBytes := int64(cfg.App().MaxFileSizeMB() * 1024 * 1024)
		if fileHeader.Size > maxSizeBytes {
			httputil.NewError(c, http.StatusRequestEntityTooLarge, fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d MB", fileHeader.Size, cfg.App().MaxFileSizeMB()))
			return
		}

		secret.Content, err = controllers.FileToBytes(fileHeader)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, err)
			return
		}
		secret.ContentType = models.ContentTypeFile
		secret.Filename = validators.SanitizeFilename(fileHeader.Filename)
	}

	if metadata, isValidationError, err := commands.CreateSecret(ctx, dataStore, encryption, secret); err != nil {
		if pkgerrors.IsContextError(err) {
			httputil.NewError(c, http.StatusRequestTimeout, err)
			return
		}
		if isValidationError {
			httputil.NewError(c, http.StatusBadRequest, err)
		} else {
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	} else if metadata == nil {
		httputil.NewError(c, http.StatusInternalServerError, errors.New("unexpected error while creating secret"))
	} else {
		c.JSON(http.StatusCreated, models.SecretMetadataResponseV2{
			ID:          metadata.ID,
			AccessCount: metadata.AccessCount,
			AccessLimit: metadata.AccessLimit,
			ContentType: metadata.ContentType,
			Filename:    metadata.Filename,
			Expiration:  metadata.Expiration,
		})
	}
}

// @Summary Access Secret Content. If the content is a file it the response will be an application/octet-stream
// @Tags v2
// @Produce application/json,application/octet-stream
// @Accept application/json
// @Param id path string true "Secret ID"
// @Success 200 {object} models.SecretContentResponse
// @Failure 404 {object} httputil.HTTPError
// @Failure 408 {object} httputil.HTTPError "Request Timeout - operation cancelled"
// @Failure 500 {object} httputil.HTTPError
// @Router /v2/secrets/{id}/access [post]
func AccessSecretContent(c *gin.Context) {
	ctx := c.Request.Context()
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)
	encryption := c.MustGet(cryptography.Key).(cryptography.Encryption)

	id := c.Param("id")

	if secret, err := commands.AccessSecret(ctx, dataStore, encryption, id); err != nil {
		if pkgerrors.IsContextError(err) {
			httputil.NewError(c, http.StatusRequestTimeout, err)
			return
		}
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	} else if secret == nil {
		c.Status(http.StatusNotFound)
	} else if secret.ContentType == models.ContentTypeFile {
		reader := bytes.NewReader(secret.Content)
		contentLength := reader.Size()
		contentType := "application/octet-stream"

		filename := secret.Filename
		if filename == "" {
			filename = fmt.Sprintf("cellar-%s", secret.ID[:8])
		}

		extraHeaders := map[string]string{
			"Content-Disposition":     fmt.Sprintf(`attachment; filename="%s"`, filename),
			"X-Content-Type-Options":  "nosniff",
			"Content-Security-Policy": "default-src 'none'",
			"X-Frame-Options":         "DENY",
			"Cache-Control":           "no-store, no-cache, must-revalidate",
		}

		c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
	} else {
		c.JSON(http.StatusOK, models.SecretContentResponse{
			ID:      secret.ID,
			Content: string(secret.Content),
		})
	}
}

// @Summary Get Secret Metadata
// @Tags v2
// @Produce json
// @Accept json
// @Param id path string true "Secret ID"
// @Success 200 {object} models.SecretMetadataResponseV2
// @Failure 404 {object} httputil.HTTPError
// @Failure 408 {object} httputil.HTTPError "Request Timeout - operation cancelled"
// @Failure 500 {object} httputil.HTTPError
// @Router /v2/secrets/{id} [get]
func GetSecretMetadata(c *gin.Context) {
	ctx := c.Request.Context()
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)

	id := c.Param("id")

	if secretMetadata := commands.GetSecretMetadata(ctx, dataStore, id); secretMetadata == nil {
		c.Status(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, models.SecretMetadataResponseV2{
			ID:          secretMetadata.ID,
			AccessCount: secretMetadata.AccessCount,
			AccessLimit: secretMetadata.AccessLimit,
			ContentType: secretMetadata.ContentType,
			Expiration:  secretMetadata.Expiration,
		})
	}
}

// @Summary Delete Secret
// @Tags v2
// @Produce json
// @Accept json
// @Param id path string true "Secret ID"
// @Success 204 ""
// @Failure 404 {object} httputil.HTTPError
// @Failure 408 {object} httputil.HTTPError "Request Timeout - operation cancelled"
// @Failure 500 {object} httputil.HTTPError
// @Router /v2/secrets/{id} [delete]
func DeleteSecret(c *gin.Context) {
	ctx := c.Request.Context()
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)

	id := c.Param("id")

	if deleted, err := commands.DeleteSecret(ctx, dataStore, id); err != nil {
		if pkgerrors.IsContextError(err) {
			httputil.NewError(c, http.StatusRequestTimeout, err)
			return
		}
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	} else if !deleted {
		c.Status(http.StatusNotFound)
	} else {
		c.Status(http.StatusNoContent)
	}
}
