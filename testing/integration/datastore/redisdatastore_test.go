// +build integration

package datastore

import (
	"cellar/pkg/datastore"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWhenGettingHealth(t *testing.T) {
	cfg := settings.NewConfiguration()
	sut := datastore.NewRedisDataStore(cfg.Redis())
	actual := sut.Health()

	t.Run("should return name", testhelpers.EqualsF("redis", strings.ToLower(actual.Name)))
	t.Run("should return healthy status", testhelpers.EqualsF("healthy", strings.ToLower(actual.Status)))
	t.Run("should return version", testhelpers.NotEqualsF("", actual.Version))
}

func TestWhenWritingSecret(t *testing.T) {
	cfg := settings.NewConfiguration()
	redisClient := testhelpers.GetRedisClient(cfg.Redis())
	sut := datastore.NewRedisDataStore(cfg.Redis())

	secret := *models.NewSecret(
		testhelpers.RandomId(t),
		testhelpers.RandomId(t),
		0,
		50,
		testhelpers.EpochFromNow(time.Minute),
	)

	keys := datastore.NewRedisKeySet(secret.ID)

	err := sut.WriteSecret(secret)

	t.Cleanup(func() {
		_ = redisClient.Del(keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	t.Run("should not return error", testhelpers.OkF(err))
	t.Run("should insert content into redis", func(t *testing.T) {
		val, err := redisClient.Get(keys.Content()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, secret.Content, val)
	})
	t.Run("should insert access limit into redis", func(t *testing.T) {
		val, err := redisClient.Get(keys.AccessLimit()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, strconv.Itoa(secret.AccessLimit), val)
	})
	t.Run("should insert access count into redis", func(t *testing.T) {
		val, err := redisClient.Get(keys.Access()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, strconv.Itoa(secret.AccessCount), val)
	})
	t.Run("should insert expiration into redis", func(t *testing.T) {
		val, err := redisClient.Get(keys.ExpirationEpoch()).Int64()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, secret.ExpirationEpoch, val)
	})
	t.Run("should set TTL on expiration", func(t *testing.T) {
		val, err := redisClient.TTL(keys.ExpirationEpoch()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		testhelpers.Ok(t, err)
		testhelpers.Assert(t, actualExpiration.Sub(secret.Expiration().Time()) <= time.Second, "Data store TTL should expire within a second of requested")
	})
	t.Run("should set TTL on access count", func(t *testing.T) {
		val, err := redisClient.TTL(keys.Access()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		testhelpers.Ok(t, err)
		testhelpers.Assert(t, actualExpiration.Sub(secret.Expiration().Time()) <= time.Second, "Data store TTL should expire within a second of requested")
	})
	t.Run("should set TTL on content", func(t *testing.T) {
		val, err := redisClient.TTL(keys.Content()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		testhelpers.Ok(t, err)
		testhelpers.Assert(t, actualExpiration.Sub(secret.Expiration().Time()) <= time.Second, "Data store TTL should expire within a second of requested")
	})
	t.Run("should set TTL on access limit", func(t *testing.T) {
		val, err := redisClient.TTL(keys.AccessLimit()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		testhelpers.Ok(t, err)
		testhelpers.Assert(t, actualExpiration.Sub(secret.Expiration().Time()) <= time.Second, "Data store TTL should expire within a second of requested")
	})
}

func TestWhenReadingSecret(t *testing.T) {
	cfg := settings.NewConfiguration()
	redis := testhelpers.GetRedisClient(cfg.Redis())
	sut := datastore.NewRedisDataStore(cfg.Redis())

	expected := *models.NewSecret(
		testhelpers.RandomId(t),
		testhelpers.RandomId(t),
		0,
		50,
		testhelpers.EpochFromNow(time.Minute),
	)

	keys := datastore.NewRedisKeySet(expected.ID)

	testhelpers.Ok(t, sut.WriteSecret(expected))

	t.Cleanup(func() {
		_ = redis.Del(keys.AllKeys()...).Err()
		_ = redis.Close()
	})

	actual := sut.ReadSecret(expected.ID)

	t.Run("should return ID", testhelpers.EqualsF(expected.ID, actual.ID))
	t.Run("should return access count", testhelpers.EqualsF(0, actual.AccessCount))
	t.Run("should return access limit", testhelpers.EqualsF(expected.AccessLimit, actual.AccessLimit))
	t.Run("should return correct expiration", testhelpers.EqualsF(expected.Expiration().Format(), actual.Expiration().Format()))
}

func TestWenDeletingSecret(t *testing.T) {
	cfg := settings.NewConfiguration()
	redis := testhelpers.GetRedisClient(cfg.Redis())
	sut := datastore.NewRedisDataStore(cfg.Redis())

	secret := *models.NewSecret(
		testhelpers.RandomId(t),
		testhelpers.RandomId(t),
		0,
		50,
		testhelpers.EpochFromNow(time.Minute),
	)

	keys := datastore.NewRedisKeySet(secret.ID)

	testhelpers.Ok(t, sut.WriteSecret(secret))

	t.Cleanup(func() {
		_ = redis.Del(keys.AllKeys()...).Err()
		_ = redis.Close()
	})

	deleted, err := sut.DeleteSecret(secret.ID)
	t.Run("should return true", testhelpers.EqualsF(true, deleted))
	t.Run("should not return error", testhelpers.OkF(err))

	t.Run("should not find content", func(t *testing.T) {
		val, err := redis.Exists(keys.Content()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
	t.Run("should not find max access", func(t *testing.T) {
		val, err := redis.Exists(keys.AccessLimit()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
	t.Run("should should not find access", func(t *testing.T) {
		val, err := redis.Exists(keys.Access()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
	t.Run("should not find expiration", func(t *testing.T) {
		val, err := redis.Exists(keys.ExpirationEpoch()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
}

func TestWhenIncreasingSecretAccess(t *testing.T) {
	cfg := settings.NewConfiguration()
	redis := testhelpers.GetRedisClient(cfg.Redis())
	sut := datastore.NewRedisDataStore(cfg.Redis())

	secret := *models.NewSecret(
		testhelpers.RandomId(t),
		testhelpers.RandomId(t),
		0,
		50,
		testhelpers.EpochFromNow(time.Minute),
	)

	keys := datastore.NewRedisKeySet(secret.ID)

	testhelpers.Ok(t, sut.WriteSecret(secret))

	t.Cleanup(func() {
		_ = redis.Del(keys.AllKeys()...).Err()
		_ = redis.Close()
	})

	actual, err := sut.IncreaseAccessCount(secret.ID)

	t.Run("should not return error", testhelpers.OkF(err))
	t.Run("should increase access count", testhelpers.EqualsF(int64(1), actual))
	t.Run("should increase access count in datastore", func(t *testing.T) {
		val, err := redis.Get(keys.Access()).Int()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, 1, val)
	})
	t.Run("should not increase access limit in datastore", func(t *testing.T) {
		val, err := redis.Get(keys.AccessLimit()).Int()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, secret.AccessLimit, val)
	})
}
