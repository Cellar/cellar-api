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

func TestWhenGettingSecret(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV1(t, cfg, content, 10)

	path := fmt.Sprintf("%s/v1/secrets/%s", cfg.App().ClientAddress(), secret.ID)
	resp, err := http.Get(path)
	require.NoError(t, err)

	t.Run("it should return ok status", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var actual models.SecretMetadataResponse
	require.NoError(t, json.Unmarshal(responseBody, &actual))

	t.Run("it should return matching id", func(t *testing.T) {
		assert.Equal(t, secret.ID, actual.ID)
	})

	t.Run("it should return matching access count", func(t *testing.T) {
		assert.Equal(t, secret.AccessCount, actual.AccessCount)
	})

	t.Run("it should return matching access limit", func(t *testing.T) {
		assert.Equal(t, secret.AccessLimit, actual.AccessLimit)
	})

	t.Run("it should return matching expiration", func(t *testing.T) {
		assert.Equal(t, secret.Expiration.Format(), actual.Expiration.Format())
	})
}

func TestWhenGettingSecretThatDoesntExist(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	path := fmt.Sprintf("%s/v1/secrets/%s", cfg.App().ClientAddress(), testhelpers.RandomId(t))
	resp, err := http.Get(path)
	require.NoError(t, err)

	t.Run("it should return not found status", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
