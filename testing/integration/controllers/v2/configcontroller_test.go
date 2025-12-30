//go:build integration
// +build integration

package v2

import (
	"cellar/pkg/controllers/v2"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("when calling the config endpoint", func(t *testing.T) {
		router := gin.New()
		cfg := settings.NewConfiguration()

		router.Use(func(c *gin.Context) {
			c.Set(settings.Key, cfg)
			c.Next()
		})

		router.GET("/v2/config", v2.GetConfig)

		req, _ := http.NewRequest("GET", "/v2/config", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Run("it should return OK status", func(t *testing.T) {
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("it should include Cache-Control header with 24 hour max-age", func(t *testing.T) {
			assert.Equal(t, "public, max-age=86400", w.Header().Get("Cache-Control"))
		})

		responseBody, err := io.ReadAll(w.Body)
		require.NoError(t, err)

		var response models.ConfigResponse
		require.NoError(t, json.Unmarshal(responseBody, &response))

		t.Run("it should return maxFileSizeMB", func(t *testing.T) {
			assert.Equal(t, cfg.App().MaxFileSizeMB(), response.Limits.MaxFileSizeMB)
		})

		t.Run("it should return maxAccessCount", func(t *testing.T) {
			assert.Equal(t, cfg.App().MaxAccessCount(), response.Limits.MaxAccessCount)
		})

		t.Run("it should return maxExpirationSeconds", func(t *testing.T) {
			assert.Equal(t, cfg.App().MaxExpirationSeconds(), response.Limits.MaxExpirationSeconds)
		})
	})
}
