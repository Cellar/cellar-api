package v2

import (
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Get Configuration
// @Tags v2
// @Produce json
// @Success 200 {object} models.ConfigResponse
// @Router /v2/config [get]
func GetConfig(c *gin.Context) {
	cfg := c.MustGet(settings.Key).(settings.IConfiguration)

	response := models.ConfigResponse{
		Limits: models.LimitsConfig{
			MaxFileSizeMB:        cfg.App().MaxFileSizeMB(),
			MaxAccessCount:       cfg.App().MaxAccessCount(),
			MaxExpirationSeconds: cfg.App().MaxExpirationSeconds(),
		},
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.JSON(http.StatusOK, response)
}
