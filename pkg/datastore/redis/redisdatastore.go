package redis

import (
	"cellar/pkg/datastore/keyvalue"
	"cellar/pkg/settings/datastore"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var Server = keyvalue.Server{Name: "Redis", VersionKey: "redis_version"}

type DataStore struct {
	*keyvalue.DataStore
	client *redis.Client
}

func NewDataStore(configuration datastore.IRedisConfiguration) *DataStore {
	address := fmt.Sprintf("%s:%d", configuration.Host(), configuration.Port())
	logger := log.WithFields(log.Fields{"context": "datastore", "instance": "redis", "address": address})
	logger.Debug("initializing redis configuration")
	if configuration.Password() == "" {
		logger.Warn("redis password is empty")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: configuration.Password(),
		DB:       configuration.DB(),
	})

	return &DataStore{
		DataStore: keyvalue.NewDataStore(clientAdapter{client}, Server),
		client:    client,
	}
}

func (store *DataStore) Client() *redis.Client {
	return store.client
}

type clientAdapter struct {
	client *redis.Client
}

func (c clientAdapter) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c clientAdapter) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c clientAdapter) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c clientAdapter) Del(ctx context.Context, keys ...string) (int64, error) {
	pipe := c.client.Pipeline()
	cmds := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Del(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	var deleted int64
	for _, cmd := range cmds {
		deleted += cmd.Val()
	}
	return deleted, nil
}

func (c clientAdapter) Info(ctx context.Context, section string) (string, error) {
	return c.client.Info(ctx, section).Result()
}

func (c clientAdapter) Close() error {
	return c.client.Close()
}
