//go:build acceptance
// +build acceptance

package config

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenGettingConfig(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	resp, err := http.Get(cfg.App().ClientAddress() + "/v2/config")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	t.Run("it should return OK status", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("it should include Cache-Control header", func(t *testing.T) {
		assert.Equal(t, "public, max-age=86400", resp.Header.Get("Cache-Control"))
	})

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var configResp models.ConfigResponse
	require.NoError(t, json.Unmarshal(responseBody, &configResp))

	t.Run("it should return maxFileSizeMB greater than zero", func(t *testing.T) {
		assert.Greater(t, configResp.Limits.MaxFileSizeMB, 0)
	})

	t.Run("it should return maxAccessCount greater than zero", func(t *testing.T) {
		assert.Greater(t, configResp.Limits.MaxAccessCount, 0)
	})

	t.Run("it should return maxExpirationSeconds greater than zero", func(t *testing.T) {
		assert.Greater(t, configResp.Limits.MaxExpirationSeconds, 0)
	})
}
