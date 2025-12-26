package middleware

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/cryptography/aws"
	"cellar/pkg/cryptography/vault"
	"cellar/pkg/datastore"
	"cellar/pkg/datastore/redis"
	"cellar/pkg/settings"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
)

func injectDependencies(router *gin.Engine, cfg settings.IConfiguration) {
	encryptionClient, err := getEncryptionClient(cfg)
	HandleError("error while initializing cryptography engine connection", err)

	dataStore := getDatastoreClient(cfg)

	router.Use(func(c *gin.Context) {
		c.Set(settings.Key, cfg)
		c.Set(cryptography.Key, encryptionClient)
		c.Set(datastore.Key, dataStore)
		c.Next()
	})
}

func getEncryptionClient(cfg settings.IConfiguration) (cryptography.Encryption, error) {
	ctx := context.Background()

	if cfg.Encryption().Vault().Enabled() {
		if cfg.Encryption().Aws().Enabled() {
			return nil, errors.New("cannot enable more than one cryptography engine")
		} else {
			return vault.NewEncryptionClient(ctx, cfg.Encryption().Vault())
		}
	} else if cfg.Encryption().Aws().Enabled() {
		return aws.NewEncryptionClient(ctx, cfg.Encryption().Aws())
	}

	return nil, errors.New("at least one cryptography engine is required")
}

func getDatastoreClient(cfg settings.IConfiguration) datastore.DataStore {
	return redis.NewDataStore(cfg.Datastore().Redis())
}
