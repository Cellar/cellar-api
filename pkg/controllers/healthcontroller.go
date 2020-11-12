package controllers

import (
	"cellar/pkg/commands"
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"github.com/gin-gonic/gin"
)

// @Summary Health Check
// @Produce  json
// @Success 200 {object} models.HealthResponse
// @Router /health-check [get]
func HealthCheck(c *gin.Context) {
	dataStore := c.MustGet(datastore.Key).(datastore.DataStore)
	encryption := c.MustGet(cryptography.Key).(cryptography.Encryption)

	health := commands.GetHealth(dataStore, encryption)
	c.JSON(200, health)
}
