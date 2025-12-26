//go:build acceptance
// +build acceptance

package secrets

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenCreatingASecretFromContent(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

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

	responseBody, err := ioutil.ReadAll(resp.Body)
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
}

func TestWhenCreatingASecretFromFile(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

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

	responseBody, err := ioutil.ReadAll(resp.Body)
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
}

func TestWhenCreatingASecretAndExpirationIsTooShort(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

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
}

func TestWhenCreatingASecretAndExpirationIsInThePast(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

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
}

func TestWhenCreatingASecretFromContentWithoutAccessLimit(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	body := map[string]string{
		"content":          "Super Secret Test Content",
		"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
	}
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

	t.Run("it should return created status", func(t *testing.T) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	responseBody, err := ioutil.ReadAll(resp.Body)
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
}

func TestWhenCreatingASecretFromFileWithoutAccessLimit(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	body := map[string]string{"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10)}
	fileContent := map[string]string{"file": "Super Secret Test Content"}
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, fileContent)

	t.Run("it should return created status", func(t *testing.T) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	responseBody, err := ioutil.ReadAll(resp.Body)
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
}

func TestWhenCreatingASecretWithoutExpiration(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	body := map[string]string{
		"content":      "Super Secret Test Content",
		"access_limit": "100",
	}
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

	t.Run("it should return bad request status", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestWhenCreatingASecretWithoutContentOrFile(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	body := map[string]string{
		"access_limit": "100",
		"duration":     strconv.FormatInt(time.Now().Add(time.Hour).UTC().Unix(), 10),
	}

	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", body, nil)

	t.Run("it should return bad request status", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestWhenCreatingASecretWithContentAndFile(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
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
}
