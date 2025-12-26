//go:build acceptance
// +build acceptance

package secrets

import (
	"bytes"
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenCreatingASecret(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	expected := struct {
		Content         string `json:"content"`
		AccessLimit     int    `json:"access_limit"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		AccessLimit:     100,
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	require.NoError(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)

	defer resp.Body.Close()

	t.Run("it should return status created", func(t *testing.T) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	responseBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var actual models.SecretMetadataResponse
	require.NoError(t, json.Unmarshal(responseBody, &actual))

	t.Run("it should have non-empty id", func(t *testing.T) {
		assert.NotEqual(t, "", actual.ID)
	})

	t.Run("it should have expected access limit", func(t *testing.T) {
		assert.Equal(t, expected.AccessLimit, actual.AccessLimit)
	})

	t.Run("it should have expected expiration", func(t *testing.T) {
		assert.Equal(t, expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format())
	})
}

func TestWhenCreatingASecretAndExpirationIsTooShort(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Minute * 9).UTC()

	expected := struct {
		Content         string `json:"content"`
		AccessLimit     int    `json:"access_limit"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		AccessLimit:     100,
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	require.NoError(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)

	defer resp.Body.Close()

	t.Run("it should return status bad request", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestWhenCreatingASecretAndExpirationIsInThePast(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour * -24).UTC()

	expected := struct {
		Content         string `json:"content"`
		AccessLimit     int    `json:"access_limit"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		AccessLimit:     100,
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	require.NoError(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)

	defer resp.Body.Close()

	t.Run("it should return status bad request", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestWhenCreatingASecretWithoutAccessLimit(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	expected := struct {
		Content         string `json:"content"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	require.NoError(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)

	defer resp.Body.Close()

	t.Run("it should return status created", func(t *testing.T) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	responseBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var actual models.SecretMetadataResponse
	require.NoError(t, json.Unmarshal(responseBody, &actual))

	t.Run("it should have non-empty id", func(t *testing.T) {
		assert.NotEqual(t, "", actual.ID)
	})

	t.Run("it should have access limit set to 0", func(t *testing.T) {
		assert.Equal(t, 0, actual.AccessLimit)
	})

	t.Run("it should have expected expiration", func(t *testing.T) {
		assert.Equal(t, expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format())
	})
}

func TestWhenCreatingASecretWithoutDuration(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	expected := map[string]interface{}{
		"content":      "Super Secret Test Content",
		"access_limit": 100,
	}
	body, err := json.Marshal(expected)
	require.NoError(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)

	defer resp.Body.Close()

	t.Run("it should return status bad request", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestWhenCreatingASecretWithoutContent(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	expected := map[string]interface{}{
		"access_limit": 100,
		"duration":     60,
	}
	body, err := json.Marshal(expected)
	require.NoError(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)

	defer resp.Body.Close()

	t.Run("it should return status bad request", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
