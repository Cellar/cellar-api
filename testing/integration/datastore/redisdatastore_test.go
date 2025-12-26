//go:build integration
// +build integration

package datastore

import (
	"cellar/pkg/datastore/redis"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenGettingHealth(t *testing.T) {
	ctx := context.Background()
	cfg := settings.NewConfiguration()
	sut := redis.NewDataStore(cfg.Datastore().Redis())
	actual := sut.Health(ctx)

	t.Run("it should return redis name", func(t *testing.T) {
		assert.Equal(t, "redis", actual.Name)
	})

	t.Run("it should return healthy status", func(t *testing.T) {
		assert.Equal(t, "healthy", actual.Status)
	})

	t.Run("it should return version", func(t *testing.T) {
		assert.NotEmpty(t, actual.Version)
	})
}

func TestWhenWritingSecret(t *testing.T) {
	ctx := context.Background()
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

	err := sut.WriteSecret(ctx, secret)

	t.Cleanup(func() {
		_ = redisClient.Del(ctx, keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	t.Run("it should not return error", func(t *testing.T) {
		assert.NoError(t, err)
	})

	t.Run("it should insert content type into redis", func(t *testing.T) {
		val, err := redisClient.Get(ctx, keys.ContentType()).Result()
		require.NoError(t, err)
		assert.Equal(t, secret.ContentType, val)
	})

	t.Run("it should insert cipher text into redis", func(t *testing.T) {
		val, err := redisClient.Get(ctx, keys.Content()).Result()
		require.NoError(t, err)
		assert.Equal(t, secret.CipherText, val)
	})

	t.Run("it should insert access limit into redis", func(t *testing.T) {
		val, err := redisClient.Get(ctx, keys.AccessLimit()).Result()
		require.NoError(t, err)
		assert.Equal(t, strconv.Itoa(secret.AccessLimit), val)
	})

	t.Run("it should insert access count into redis", func(t *testing.T) {
		val, err := redisClient.Get(ctx, keys.Access()).Result()
		require.NoError(t, err)
		assert.Equal(t, strconv.Itoa(secret.AccessCount), val)
	})

	t.Run("it should insert expiration into redis", func(t *testing.T) {
		val, err := redisClient.Get(ctx, keys.ExpirationEpoch()).Int64()
		require.NoError(t, err)
		assert.Equal(t, secret.ExpirationEpoch, val)
	})

	t.Run("it should set TTL on expiration", func(t *testing.T) {
		val, err := redisClient.TTL(ctx, keys.ExpirationEpoch()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		require.NoError(t, err)
		assert.LessOrEqual(t, actualExpiration.Sub(secret.Expiration().Time()), time.Second)
	})

	t.Run("it should set TTL on access count", func(t *testing.T) {
		val, err := redisClient.TTL(ctx, keys.Access()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		require.NoError(t, err)
		assert.LessOrEqual(t, actualExpiration.Sub(secret.Expiration().Time()), time.Second)
	})

	t.Run("it should set TTL on content type", func(t *testing.T) {
		val, err := redisClient.TTL(ctx, keys.ContentType()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		require.NoError(t, err)
		assert.LessOrEqual(t, actualExpiration.Sub(secret.Expiration().Time()), time.Second)
	})

	t.Run("it should set TTL on content", func(t *testing.T) {
		val, err := redisClient.TTL(ctx, keys.Content()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		require.NoError(t, err)
		assert.LessOrEqual(t, actualExpiration.Sub(secret.Expiration().Time()), time.Second)
	})

	t.Run("it should set TTL on access limit", func(t *testing.T) {
		val, err := redisClient.TTL(ctx, keys.AccessLimit()).Result()
		actualExpiration := time.Now().Add(val).UTC()
		require.NoError(t, err)
		assert.LessOrEqual(t, actualExpiration.Sub(secret.Expiration().Time()), time.Second)
	})
}

func TestWhenReadingSecret(t *testing.T) {
	ctx := context.Background()
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

	require.NoError(t, sut.WriteSecret(ctx, expected))

	t.Cleanup(func() {
		_ = redisClient.Del(ctx, keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	actual := sut.ReadSecret(ctx, expected.ID)

	t.Run("it should return ID", func(t *testing.T) {
		assert.Equal(t, expected.ID, actual.ID)
	})

	t.Run("it should return content", func(t *testing.T) {
		assert.Equal(t, expected.CipherText, actual.CipherText)
	})

	t.Run("it should return content type", func(t *testing.T) {
		assert.Equal(t, expected.ContentType, actual.ContentType)
	})

	t.Run("it should return access count", func(t *testing.T) {
		assert.Equal(t, 0, actual.AccessCount)
	})

	t.Run("it should return access limit", func(t *testing.T) {
		assert.Equal(t, expected.AccessLimit, actual.AccessLimit)
	})

	t.Run("it should return correct expiration", func(t *testing.T) {
		assert.Equal(t, expected.Expiration().Format(), actual.Expiration().Format())
	})
}

func TestWenDeletingSecret(t *testing.T) {
	ctx := context.Background()
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

	require.NoError(t, sut.WriteSecret(ctx, secret))

	t.Cleanup(func() {
		_ = redisClient.Del(ctx, keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	deleted, err := sut.DeleteSecret(ctx, secret.ID)

	t.Run("it should return true", func(t *testing.T) {
		assert.True(t, deleted)
	})

	t.Run("it should not return error", func(t *testing.T) {
		assert.NoError(t, err)
	})

	t.Run("it should not find content type", func(t *testing.T) {
		val, err := redisClient.Exists(ctx, keys.ContentType()).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), val)
	})

	t.Run("it should not find content", func(t *testing.T) {
		val, err := redisClient.Exists(ctx, keys.Content()).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), val)
	})

	t.Run("it should not find max access", func(t *testing.T) {
		val, err := redisClient.Exists(ctx, keys.AccessLimit()).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), val)
	})

	t.Run("it should not find access", func(t *testing.T) {
		val, err := redisClient.Exists(ctx, keys.Access()).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), val)
	})

	t.Run("it should not find expiration", func(t *testing.T) {
		val, err := redisClient.Exists(ctx, keys.ExpirationEpoch()).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), val)
	})
}

func TestWhenIncreasingSecretAccess(t *testing.T) {
	ctx := context.Background()
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

	require.NoError(t, sut.WriteSecret(ctx, secret))

	t.Cleanup(func() {
		_ = redisClient.Del(ctx, keys.AllKeys()...).Err()
		_ = redisClient.Close()
	})

	actual, err := sut.IncreaseAccessCount(ctx, secret.ID)

	t.Run("it should not return error", func(t *testing.T) {
		assert.NoError(t, err)
	})

	t.Run("it should increase access count", func(t *testing.T) {
		assert.Equal(t, int64(1), actual)
	})

	t.Run("it should increase access count in datastore", func(t *testing.T) {
		val, err := redisClient.Get(ctx, keys.Access()).Int()
		require.NoError(t, err)
		assert.Equal(t, 1, val)
	})

	t.Run("it should not increase access limit in datastore", func(t *testing.T) {
		val, err := redisClient.Get(ctx, keys.AccessLimit()).Int()
		require.NoError(t, err)
		assert.Equal(t, secret.AccessLimit, val)
	})
}
