package v2

import (
	"cellar/pkg/commands"
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

	limits := commands.GetConfig(cfg)

	response := models.ConfigResponse{
		Limits: limits,
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.JSON(http.StatusOK, response)
}
