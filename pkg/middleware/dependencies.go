package middleware

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/cryptography/aws"
	"cellar/pkg/cryptography/vault"
	"cellar/pkg/datastore"
	"cellar/pkg/datastore/redis"
	"cellar/pkg/datastore/valkey"
	"cellar/pkg/ratelimit"
	"cellar/pkg/settings"
	datastoresettings "cellar/pkg/settings/datastore"
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

type datastoreKind int

const (
	redisDatastore datastoreKind = iota + 1
	valkeyDatastore
)

func injectDependencies(router *gin.Engine, cfg settings.IConfiguration) {
	encryptionClient, err := getEncryptionClient(cfg)
	HandleError("error while initializing cryptography engine connection", err)

	dataStore, windowStore, err := getDatastoreClient(cfg)
	HandleError("error while initializing datastore connection", err)

	rateLimiter := ratelimit.NewSlidingWindowRateLimiter(windowStore, cfg.RateLimit())

	router.Use(func(c *gin.Context) {
		c.Set(settings.Key, cfg)
		c.Set(cryptography.Key, encryptionClient)
		c.Set(datastore.Key, dataStore)
		c.Set(ratelimit.Key, rateLimiter)
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

func selectDatastore(cfg datastoresettings.IDatastoreConfiguration) (datastoreKind, error) {
	redisEnabled := cfg.Redis().Enabled()
	valkeyEnabled := cfg.Valkey().Enabled()

	switch {
	case redisEnabled && valkeyEnabled:
		return 0, errors.New("cannot enable more than one datastore")
	case redisEnabled:
		return redisDatastore, nil
	case valkeyEnabled:
		return valkeyDatastore, nil
	default:
		return 0, errors.New("at least one datastore is required")
	}
}

func getDatastoreClient(cfg settings.IConfiguration) (datastore.DataStore, ratelimit.WindowStore, error) {
	kind, err := selectDatastore(cfg.Datastore())
	if err != nil {
		return nil, nil, err
	}

	if kind == valkeyDatastore {
		store, err := valkey.NewDataStore(cfg.Datastore().Valkey())
		if err != nil {
			return nil, nil, err
		}
		return store, ratelimit.NewValkeyWindowStore(store.Client()), nil
	}

	store := redis.NewDataStore(cfg.Datastore().Redis())
	return store, ratelimit.NewRedisWindowStore(store.Client()), nil
}
