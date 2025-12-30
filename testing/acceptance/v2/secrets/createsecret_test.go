//go:build acceptance
// +build acceptance

package secrets

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSecret(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	t.Run("when creating a secret from content", func(t *testing.T) {
		t.Run("and all parameters are valid", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Hour).UTC()
			expectedAccessLimit := 100

			body := map[string]string{
				"content":          "Super Secret Test Content",
				"access_limit":     strconv.Itoa(expectedAccessLimit),
				"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
			}

			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return created status", func(t *testing.T) {
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
			})

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var actual models.SecretMetadataResponseV2
			require.NoError(t, json.Unmarshal(responseBody, &actual))

			t.Run("it should have non-empty id", func(t *testing.T) {
				assert.NotEqual(t, "", actual.ID)
			})

			t.Run("it should set access limit", func(t *testing.T) {
				assert.Equal(t, expectedAccessLimit, actual.AccessLimit)
			})

			t.Run("it should set expiration", func(t *testing.T) {
				assert.Equal(t, expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format())
			})

			t.Run("it should have text content type", func(t *testing.T) {
				assert.Equal(t, "text", string(actual.ContentType))
			})
		})

		t.Run("and access limit is not provided", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Hour).UTC()

			body := map[string]string{
				"content":          "Super Secret Test Content",
				"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
			}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return created status", func(t *testing.T) {
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
			})

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var actual models.SecretMetadataResponseV2
			require.NoError(t, json.Unmarshal(responseBody, &actual))

			t.Run("it should have non-empty id", func(t *testing.T) {
				assert.NotEqual(t, "", actual.ID)
			})

			t.Run("it should set access limit to 0", func(t *testing.T) {
				assert.Equal(t, 0, actual.AccessLimit)
			})

			t.Run("it should set expiration", func(t *testing.T) {
				assert.Equal(t, expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format())
			})

			t.Run("it should have text content type", func(t *testing.T) {
				assert.Equal(t, "text", string(actual.ContentType))
			})
		})

		t.Run("and access limit exceeds maximum configured value", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Hour).UTC()
			maxAccessCount := cfg.App().MaxAccessCount()
			exceedingAccessLimit := maxAccessCount + 1

			body := map[string]string{
				"content":          "Super Secret Test Content",
				"access_limit":     strconv.Itoa(exceedingAccessLimit),
				"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
			}

			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})

		t.Run("and expiration is not provided", func(t *testing.T) {
			body := map[string]string{
				"content":      "Super Secret Test Content",
				"access_limit": "100",
			}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})

		t.Run("and expiration is in the past", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Hour * -24).UTC()

			body := map[string]string{
				"content":          "Super Secret Test Content",
				"access_limit":     "100",
				"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
			}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})

		t.Run("and expiration is too short", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Minute * 9).UTC()

			body := map[string]string{
				"content":          "Super Secret Test Content",
				"access_limit":     "100",
				"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
			}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})

		t.Run("and expiration exceeds maximum configured value", func(t *testing.T) {
			maxExpirationSeconds := cfg.App().MaxExpirationSeconds()
			exceedingExpiration := time.Now().Add(time.Second * time.Duration(maxExpirationSeconds+1)).UTC()

			body := map[string]string{
				"content":          "Super Secret Test Content",
				"access_limit":     "10",
				"expiration_epoch": strconv.FormatInt(exceedingExpiration.Unix(), 10),
			}

			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})

		t.Run("and content is not provided", func(t *testing.T) {
			body := map[string]string{
				"access_limit": "100",
				"duration":     strconv.FormatInt(time.Now().Add(time.Hour).UTC().Unix(), 10),
			}

			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})

		t.Run("and both content and file are provided", func(t *testing.T) {
			body := map[string]string{
				"content":      "Super Secret Test Content",
				"access_limit": "100",
				"duration":     strconv.FormatInt(time.Now().Add(time.Hour).UTC().Unix(), 10),
			}
			fileContent := map[string]string{"file": "Super Secret Test Content"}

			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, fileContent)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})
	})

	t.Run("when creating a secret from file", func(t *testing.T) {
		t.Run("and all parameters are valid", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Hour).UTC()
			expectedAccessLimit := 100

			body := map[string]string{
				"access_limit":     strconv.Itoa(expectedAccessLimit),
				"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
			}

			fileContent := map[string]string{"file": "Super Secret Test Content"}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, fileContent)

			t.Run("it should return created status", func(t *testing.T) {
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
			})

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var actual models.SecretMetadataResponseV2
			require.NoError(t, json.Unmarshal(responseBody, &actual))

			t.Run("it should have non-empty id", func(t *testing.T) {
				assert.NotEqual(t, "", actual.ID)
			})

			t.Run("it should set access limit", func(t *testing.T) {
				assert.Equal(t, expectedAccessLimit, actual.AccessLimit)
			})

			t.Run("it should set expiration", func(t *testing.T) {
				assert.Equal(t, expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format())
			})

			t.Run("it should have file content type", func(t *testing.T) {
				assert.Equal(t, "file", string(actual.ContentType))
			})
		})

		t.Run("and access limit is not provided", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Hour).UTC()

			body := map[string]string{"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10)}
			fileContent := map[string]string{"file": "Super Secret Test Content"}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, fileContent)

			t.Run("it should return created status", func(t *testing.T) {
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
			})

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var actual models.SecretMetadataResponseV2
			require.NoError(t, json.Unmarshal(responseBody, &actual))

			t.Run("it should have non-empty id", func(t *testing.T) {
				assert.NotEqual(t, "", actual.ID)
			})

			t.Run("it should set access limit to 0", func(t *testing.T) {
				assert.Equal(t, 0, actual.AccessLimit)
			})

			t.Run("it should set expiration", func(t *testing.T) {
				assert.Equal(t, expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format())
			})

			t.Run("it should have file content type", func(t *testing.T) {
				assert.Equal(t, "file", string(actual.ContentType))
			})
		})

		t.Run("and access limit exceeds maximum configured value", func(t *testing.T) {
			expectedExpiration := time.Now().Add(time.Hour).UTC()
			maxAccessCount := cfg.App().MaxAccessCount()
			exceedingAccessLimit := maxAccessCount + 1

			body := map[string]string{
				"access_limit":     strconv.Itoa(exceedingAccessLimit),
				"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
			}

			fileContent := map[string]string{"file": "Super Secret Test Content"}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, fileContent)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})

		t.Run("and expiration exceeds maximum configured value", func(t *testing.T) {
			maxExpirationSeconds := cfg.App().MaxExpirationSeconds()
			exceedingExpiration := time.Now().Add(time.Second * time.Duration(maxExpirationSeconds+1)).UTC()

			body := map[string]string{
				"access_limit":     "10",
				"expiration_epoch": strconv.FormatInt(exceedingExpiration.Unix(), 10),
			}

			fileContent := map[string]string{"file": "Super Secret Test Content"}
			resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, fileContent)

			t.Run("it should return bad request status", func(t *testing.T) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		})
	})
}
