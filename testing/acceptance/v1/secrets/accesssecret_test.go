//go:build acceptance
// +build acceptance

package secrets

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenAccessingSecretContent(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV1(t, cfg, content, 10)

	path := fmt.Sprintf("%s/v1/secrets/%s/access", cfg.App().ClientAddress(), secret.ID)
	resp, err := http.Post(path, "application/json", nil)
	require.NoError(t, err)

	t.Run("it should return ok status", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var actual models.SecretContentResponse
	require.NoError(t, json.Unmarshal(responseBody, &actual))

	t.Run("it should have matching id", func(t *testing.T) {
		assert.Equal(t, secret.ID, actual.ID)
	})

	t.Run("it should have matching content", func(t *testing.T) {
		assert.Equal(t, content, actual.Content)
	})
}

func TestWhenAccessingSecretContentForSecretThatDoesntExist(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	path := fmt.Sprintf("%s/v1/secrets/%s/content", cfg.App().ClientAddress(), testhelpers.RandomId(t))
	resp, err := http.Get(path)
	require.NoError(t, err)

	t.Run("it should return not found status", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestWhenAccessingSecretWithAccessLimitOfOne(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV1(t, cfg, content, 1)

	path := fmt.Sprintf("%s/v1/secrets/%s/access", cfg.App().ClientAddress(), secret.ID)
	response1, err := http.Post(path, "application/json", nil)
	require.NoError(t, err)
	response2, err := http.Post(path, "application/json", nil)
	require.NoError(t, err)

	t.Run("it should return ok status for first request", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, response1.StatusCode)
	})

	t.Run("it should return not found status for second request", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, response2.StatusCode)
	})

	defer response1.Body.Close()

	responseBody, err := ioutil.ReadAll(response1.Body)
	require.NoError(t, err)

	var actual models.SecretContentResponse
	require.NoError(t, json.Unmarshal(responseBody, &actual))

	t.Run("it should have matching id in response1", func(t *testing.T) {
		assert.Equal(t, secret.ID, actual.ID)
	})

	t.Run("it should have matching content in response1", func(t *testing.T) {
		assert.Equal(t, content, actual.Content)
	})
}
