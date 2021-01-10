package middleware

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/cryptography/vault"
	"cellar/pkg/datastore"
	"cellar/pkg/datastore/redis"
	"cellar/pkg/settings"
	"github.com/gin-gonic/gin"
)

func injectDependencies(router *gin.Engine, cfg settings.IConfiguration) {
	vaultEncryptionClient, err := vault.NewEncryptionClient(cfg.Vault())
	HandleError("error while initializing vault connection", err)
	dataStore := redis.NewDataStore(cfg.Redis())

	router.Use(func(c *gin.Context) {
		c.Set(settings.Key, cfg)
		c.Set(cryptography.Key, vaultEncryptionClient)
		c.Set(datastore.Key, dataStore)
		c.Next()
	})
}
