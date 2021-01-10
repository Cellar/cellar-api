package datastore

import (
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"fmt"
	"github.com/go-redis/redis/v7"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type RedisDataStore struct {
	client *redis.Client
	logger *log.Entry
}

type RedisInfo struct {
	Version string `json:"redis_version"`
}

const redisIdFieldKey = "redis_key"

func NewRedisDataStore(configuration settings.IRedisConfiguration) *RedisDataStore {

	return &RedisDataStore{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", configuration.Host(), configuration.Port()),
			Password: configuration.Password(),
			DB:       configuration.DB(),
		}),
		logger: initializeLogger(configuration),
	}
}

func initializeLogger(configuration settings.IRedisConfiguration) *log.Entry {
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

func (redis RedisDataStore) Health() models.Health {
	name := "Redis"
	status := models.HealthStatus(models.Unhealthy)
	version := "Unknown"

	if res, err := redis.client.Info("server").Result(); err == nil {
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

func (redis RedisDataStore) WriteSecret(secret models.Secret) error {
	keySet := NewRedisKeySet(secret.ID)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("Writing secret to datastore")

	err := redis.client.Set(keySet.Access(), strconv.Itoa(0), secret.Duration()).Err()
	if err != nil {
		return err
	}
	err = redis.client.Set(keySet.AccessLimit(), strconv.Itoa(secret.AccessLimit), secret.Duration()).Err()
	if err != nil {
		return err
	}
	err = redis.client.Set(keySet.Content(), secret.Content, secret.Duration()).Err()
	if err != nil {
		return err
	}
	err = redis.client.Set(keySet.ExpirationEpoch(), secret.ExpirationEpoch, secret.Duration()).Err()
	if err != nil {
		return err
	}

	return nil
}

func (redis RedisDataStore) ReadSecret(id string) (secret *models.Secret) {
	keySet := NewRedisKeySet(id)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("reading secret from redis")

	accessLimit, err := redis.client.Get(keySet.AccessLimit()).Int()
	if err != nil {
		return nil
	}

	content, err := redis.client.Get(keySet.Content()).Result()
	if err != nil {
		return nil
	}

	accessCount, err := redis.client.Get(keySet.Access()).Int()
	if err != nil {
		return nil
	}

	expirationEpoch, err := redis.client.Get(keySet.ExpirationEpoch()).Int64()
	if err != nil {
		return nil
	}

	return models.NewSecret(
		id,
		content,
		accessCount,
		accessLimit,
		expirationEpoch,
	)
}

func (redis RedisDataStore) IncreaseAccessCount(id string) (accessCount int64, err error) {
	keySet := NewRedisKeySet(id)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("increasing secret access count in redis")
	return redis.client.Incr(keySet.Access()).Result()
}

func (redis RedisDataStore) DeleteSecret(id string) (bool, error) {
	keySet := NewRedisKeySet(id)
	redis.logger.WithField(redisIdFieldKey, keySet.id).Debug("deleting secret from redis")
	numDeleted, err := redis.client.Del(keySet.AllKeys()...).Result()
	return numDeleted > int64(0), err
}

func (redis RedisDataStore) Close() error {
	return redis.client.Close()
}
