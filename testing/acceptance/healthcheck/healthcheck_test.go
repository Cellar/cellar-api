//go:build acceptance
// +build acceptance

package healthcheck

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	resp, err := http.Get(cfg.App().ClientAddress() + "/health-check")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	t.Run("it should return ok status", func(t *testing.T) {
		assert.Equal(t, 200, resp.StatusCode)
	})

	responseBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var health models.HealthResponse
	require.NoError(t, json.Unmarshal(responseBody, &health))

	t.Run("it should return healthy status", func(t *testing.T) {
		assert.Equal(t, "healthy", strings.ToLower(health.Status))
	})

	t.Run("it should return non-empty host", func(t *testing.T) {
		assert.NotEqual(t, "", health.Host)
	})

	t.Run("it should return non-empty version", func(t *testing.T) {
		assert.NotEqual(t, "", health.Version)
	})

	t.Run("it should return redis datastore name", func(t *testing.T) {
		assert.Equal(t, "redis", strings.ToLower(health.Datastore.Name))
	})

	t.Run("it should return healthy datastore status", func(t *testing.T) {
		assert.Equal(t, "healthy", strings.ToLower(health.Datastore.Status))
	})

	t.Run("it should return non-empty datastore version", func(t *testing.T) {
		assert.NotEqual(t, "", health.Datastore.Version)
	})

	t.Run("it should return vault encryption name", func(t *testing.T) {
		assert.Equal(t, "vault", strings.ToLower(health.Encryption.Name))
	})

	t.Run("it should return healthy encryption status", func(t *testing.T) {
		assert.Equal(t, "healthy", strings.ToLower(health.Encryption.Status))
	})

	t.Run("it should return non-empty encryption version", func(t *testing.T) {
		assert.NotEqual(t, "", health.Encryption.Version)
	})
}
