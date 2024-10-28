package controllers

import (
	"cellar/pkg/commands"
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/settings"
	"github.com/gin-gonic/gin"
)

// @Summary Health Check
// @Tags common
// @Produce  json
// @Success 200 {object} models.HealthResponse
// @Router /health-check [get]
func HealthCheck(c *gin.Context) {
	cfg := c.MustGet(settings.Key).(settings.IConfiguration)
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)
	encryption := c.MustGet(cryptography.Key).(cryptography.Encryption)

	health := commands.GetHealth(cfg.App(), dataStore, encryption)
	c.JSON(200, health)
}
