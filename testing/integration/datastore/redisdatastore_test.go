//go:build integration
// +build integration

package datastore

import (
	"cellar/pkg/datastore/redis"
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
	sut := redis.NewDataStore(cfg.Datastore().Redis())
	actual := sut.Health()

	t.Run("should return name", testhelpers.EqualsF("redis", strings.ToLower(actual.Name)))
	t.Run("should return healthy status", testhelpers.EqualsF("healthy", strings.ToLower(actual.Status)))
	t.Run("should return version", testhelpers.NotEqualsF("", actual.Version))
}

func TestWhenWritingSecret(t *testing.T) {
	cfg := settings.NewConfiguration()
	redisClient := testhelpers.GetRedisClient(cfg.Datastore().Redis())
	sut := redis.NewDataStore(cfg.Datastore().Redis())

	secret := models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      testhelpers.RandomId(t),
		ContentType:     models.ContentTypeText,
		AccessLimit:     50,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	keys := redis.NewRedisKeySet(secret.ID)

	err := sut.WriteSecret(secret)

	t.Cleanup(func() {
		_ = redisClient.Del(keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	t.Run("should not return error", testhelpers.OkF(err))
	t.Run("should insert content type into redis", func(t *testing.T) {
		val, err := redisClient.Get(keys.ContentType()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, secret.ContentType, val)
	})
	t.Run("should insert cipher text into redis", func(t *testing.T) {
		val, err := redisClient.Get(keys.Content()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, secret.CipherText, val)
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
	t.Run("should set TTL on content type", func(t *testing.T) {
		val, err := redisClient.TTL(keys.ContentType()).Result()
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
	redisClient := testhelpers.GetRedisClient(cfg.Datastore().Redis())
	sut := redis.NewDataStore(cfg.Datastore().Redis())

	expected := models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      testhelpers.RandomId(t),
		ContentType:     models.ContentTypeText,
		AccessLimit:     50,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	keys := redis.NewRedisKeySet(expected.ID)

	testhelpers.Ok(t, sut.WriteSecret(expected))

	t.Cleanup(func() {
		_ = redisClient.Del(keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	actual := sut.ReadSecret(expected.ID)

	t.Run("should return ID", testhelpers.EqualsF(expected.ID, actual.ID))
	t.Run("should return content", testhelpers.EqualsF(expected.CipherText, actual.CipherText))
	t.Run("should return content type", testhelpers.EqualsF(expected.ContentType, actual.ContentType))
	t.Run("should return access count", testhelpers.EqualsF(0, actual.AccessCount))
	t.Run("should return access limit", testhelpers.EqualsF(expected.AccessLimit, actual.AccessLimit))
	t.Run("should return correct expiration", testhelpers.EqualsF(expected.Expiration().Format(), actual.Expiration().Format()))
}

func TestWenDeletingSecret(t *testing.T) {
	cfg := settings.NewConfiguration()
	redisClient := testhelpers.GetRedisClient(cfg.Datastore().Redis())
	sut := redis.NewDataStore(cfg.Datastore().Redis())

	secret := models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      models.ContentTypeText,
		ContentType:     testhelpers.RandomId(t),
		AccessLimit:     50,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	keys := redis.NewRedisKeySet(secret.ID)

	testhelpers.Ok(t, sut.WriteSecret(secret))

	t.Cleanup(func() {
		_ = redisClient.Del(keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	deleted, err := sut.DeleteSecret(secret.ID)
	t.Run("should return true", testhelpers.EqualsF(true, deleted))
	t.Run("should not return error", testhelpers.OkF(err))

	t.Run("should not find content type", func(t *testing.T) {
		val, err := redisClient.Exists(keys.ContentType()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
	t.Run("should not find content", func(t *testing.T) {
		val, err := redisClient.Exists(keys.Content()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
	t.Run("should not find max access", func(t *testing.T) {
		val, err := redisClient.Exists(keys.AccessLimit()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
	t.Run("should should not find access", func(t *testing.T) {
		val, err := redisClient.Exists(keys.Access()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
	t.Run("should not find expiration", func(t *testing.T) {
		val, err := redisClient.Exists(keys.ExpirationEpoch()).Result()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, int64(0), val)
	})
}

func TestWhenIncreasingSecretAccess(t *testing.T) {
	cfg := settings.NewConfiguration()
	redisClient := testhelpers.GetRedisClient(cfg.Datastore().Redis())
	sut := redis.NewDataStore(cfg.Datastore().Redis())

	secret := models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      models.ContentTypeText,
		ContentType:     testhelpers.RandomId(t),
		AccessLimit:     50,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	keys := redis.NewRedisKeySet(secret.ID)

	testhelpers.Ok(t, sut.WriteSecret(secret))

	t.Cleanup(func() {
		_ = redisClient.Del(keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	actual, err := sut.IncreaseAccessCount(secret.ID)

	t.Run("should not return error", testhelpers.OkF(err))
	t.Run("should increase access count", testhelpers.EqualsF(int64(1), actual))
	t.Run("should increase access count in datastore", func(t *testing.T) {
		val, err := redisClient.Get(keys.Access()).Int()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, 1, val)
	})
	t.Run("should not increase access limit in datastore", func(t *testing.T) {
		val, err := redisClient.Get(keys.AccessLimit()).Int()
		testhelpers.Ok(t, err)
		testhelpers.Equals(t, secret.AccessLimit, val)
	})
}
