package redis

import (
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/models"
	"cellar/pkg/settings/datastore"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type (
	DataStore struct {
		client *redis.Client
		logger *log.Entry
	}
	Info struct {
		Version string `json:"redis_version"`
	}
)

const redisIdFieldKey = "redis_key"

func NewDataStore(configuration datastore.IRedisConfiguration) *DataStore {

	return &DataStore{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", configuration.Host(), configuration.Port()),
			Password: configuration.Password(),
			DB:       configuration.DB(),
		}),
		logger: initializeLogger(configuration),
	}
}

func initializeLogger(configuration datastore.IRedisConfiguration) *log.Entry {
	logger := log.WithFields(log.Fields{
		"context":  "datastore",
		"instance": "redis",
		"address":  fmt.Sprintf("%s:%d", configuration.Host(), configuration.Port()),
	})

	logger.Debug("initializing redis configuration")
	if configuration.Password() == "" {
		logger.Warn("redis password is empty")
	}

	return logger
}

func (redis DataStore) Health(ctx context.Context) models.Health {
	name := "Redis"
	status := models.HealthStatus(models.Unhealthy)
	version := "Unknown"

	if err := pkgerrors.CheckContext(ctx); err != nil {
		return *models.NewHealth(name, status, version)
	}

	if res, err := redis.client.Info(ctx, "server").Result(); err == nil {
		res = strings.ReplaceAll(res, "\r\n", "\n")
		info := strings.Split(res, "\n")
		statusKey := "redis_version"
		for _, line := range info {
			if strings.Contains(line, statusKey) {
				status = models.Healthy
				version = line[len(statusKey)+1:]
				break
			}
		}
	}

	return *models.NewHealth(name, status, version)
}

func (redis DataStore) WriteSecret(ctx context.Context, secret models.Secret) error {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return err
	}

	keySet := NewRedisKeySet(secret.ID)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("Writing secret to datastore")

	err := redis.client.Set(ctx, keySet.Access(), strconv.Itoa(0), secret.Duration()).Err()
	if err != nil {
		return err
	}
	err = redis.client.Set(ctx, keySet.AccessLimit(), strconv.Itoa(secret.AccessLimit), secret.Duration()).Err()
	if err != nil {
		return err
	}
	err = redis.client.Set(ctx, keySet.ContentType(), secret.ContentType, secret.Duration()).Err()
	if err != nil {
		return err
	}
	err = redis.client.Set(ctx, keySet.Content(), secret.CipherText, secret.Duration()).Err()
	if err != nil {
		return err
	}
	err = redis.client.Set(ctx, keySet.ExpirationEpoch(), secret.ExpirationEpoch, secret.Duration()).Err()
	if err != nil {
		return err
	}

	return nil
}

func (redis DataStore) ReadSecret(ctx context.Context, id string) (secret *models.Secret) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil
	}

	keySet := NewRedisKeySet(id)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("reading secret from redis")

	accessLimit, err := redis.client.Get(ctx, keySet.AccessLimit()).Int()
	if err != nil {
		return nil
	}

	contentType, err := redis.client.Get(ctx, keySet.ContentType()).Result()
	if err != nil {
		return nil
	}

	content, err := redis.client.Get(ctx, keySet.Content()).Result()
	if err != nil {
		return nil
	}

	accessCount, err := redis.client.Get(ctx, keySet.Access()).Int()
	if err != nil {
		return nil
	}

	expirationEpoch, err := redis.client.Get(ctx, keySet.ExpirationEpoch()).Int64()
	if err != nil {
		return nil
	}

	return &models.Secret{
		ID:              id,
		CipherText:      content,
		ContentType:     contentType,
		AccessCount:     accessCount,
		AccessLimit:     accessLimit,
		ExpirationEpoch: expirationEpoch,
	}
}

func (redis DataStore) IncreaseAccessCount(ctx context.Context, id string) (accessCount int64, err error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return 0, err
	}
	keySet := NewRedisKeySet(id)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("increasing secret access count in redis")
	return redis.client.Incr(ctx, keySet.Access()).Result()
}

func (redis DataStore) DeleteSecret(ctx context.Context, id string) (bool, error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return false, err
	}
	keySet := NewRedisKeySet(id)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("deleting secret from redis")
	numDeleted, err := redis.client.Del(ctx, keySet.AllKeys()...).Result()
	return numDeleted > int64(0), err
}

func (redis DataStore) Close() error {
	return redis.client.Close()
}
