package middleware

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/settings"
	"github.com/gin-gonic/gin"
)

func injectDependencies(router *gin.Engine, cfg settings.IConfiguration) {
	vault, err := cryptography.NewVaultEncryption(cfg)
	HandleError("error while initializing vault connection", err)
	dataStore := datastore.NewRedisDataStore(cfg)

	router.Use(func(c *gin.Context) {
		c.Set(settings.Key, cfg)
		c.Set(cryptography.Key, vault)
		c.Set(datastore.Key, dataStore)
		c.Next()
	})
}
