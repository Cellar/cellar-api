//go:build integration
// +build integration

package datastore

import (
	"cellar/pkg/datastore/redis"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"context"
	"fmt"
	"strconv"
	"strings"
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
		assert.True(t, strings.EqualFold("redis", actual.Name), "expected name to be 'redis' (case insensitive), got %s", actual.Name)
	})

	t.Run("it should return healthy status", func(t *testing.T) {
		assert.True(t, strings.EqualFold("healthy", actual.Status), "expected status to be 'healthy' (case insensitive), got %s", actual.Status)
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

	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	var testCases = []struct {
		name     string
		filename string
	}{
		{name: "has filename", filename: "test-file.pdf"},
		{name: "does not have filename", filename: ""},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("and secret %s", tc.name), func(t *testing.T) {
			testSecret := secret
			testSecret.ID = testhelpers.RandomId(t)
			testSecret.Filename = tc.filename

			keys := redis.NewRedisKeySet(testSecret.ID)
			err := sut.WriteSecret(ctx, testSecret)

			t.Cleanup(func() {
				_ = redisClient.Del(ctx, keys.AllKeys()...).Err()
			})

			t.Run("it should not return error", func(t *testing.T) {
				assert.NoError(t, err)
			})

			t.Run("it should insert content type into redis", func(t *testing.T) {
				val, err := redisClient.Get(ctx, keys.ContentType()).Result()
				require.NoError(t, err)
				assert.Equal(t, testSecret.ContentType, val)
			})

			t.Run("it should insert cipher text into redis", func(t *testing.T) {
				val, err := redisClient.Get(ctx, keys.Content()).Result()
				require.NoError(t, err)
				assert.Equal(t, testSecret.CipherText, val)
			})

			t.Run("it should insert access limit into redis", func(t *testing.T) {
				val, err := redisClient.Get(ctx, keys.AccessLimit()).Result()
				require.NoError(t, err)
				assert.Equal(t, strconv.Itoa(testSecret.AccessLimit), val)
			})

			t.Run("it should insert access count into redis", func(t *testing.T) {
				val, err := redisClient.Get(ctx, keys.Access()).Result()
				require.NoError(t, err)
				assert.Equal(t, strconv.Itoa(testSecret.AccessCount), val)
			})

			t.Run("it should insert expiration into redis", func(t *testing.T) {
				val, err := redisClient.Get(ctx, keys.ExpirationEpoch()).Int64()
				require.NoError(t, err)
				assert.Equal(t, testSecret.ExpirationEpoch, val)
			})

			t.Run("it should set TTL on expiration", func(t *testing.T) {
				val, err := redisClient.TTL(ctx, keys.ExpirationEpoch()).Result()
				actualExpiration := time.Now().Add(val).UTC()
				require.NoError(t, err)
				assert.LessOrEqual(t, actualExpiration.Sub(testSecret.Expiration().Time()), time.Second)
			})

			t.Run("it should set TTL on access count", func(t *testing.T) {
				val, err := redisClient.TTL(ctx, keys.Access()).Result()
				actualExpiration := time.Now().Add(val).UTC()
				require.NoError(t, err)
				assert.LessOrEqual(t, actualExpiration.Sub(testSecret.Expiration().Time()), time.Second)
			})

			t.Run("it should set TTL on content type", func(t *testing.T) {
				val, err := redisClient.TTL(ctx, keys.ContentType()).Result()
				actualExpiration := time.Now().Add(val).UTC()
				require.NoError(t, err)
				assert.LessOrEqual(t, actualExpiration.Sub(testSecret.Expiration().Time()), time.Second)
			})

			t.Run("it should set TTL on content", func(t *testing.T) {
				val, err := redisClient.TTL(ctx, keys.Content()).Result()
				actualExpiration := time.Now().Add(val).UTC()
				require.NoError(t, err)
				assert.LessOrEqual(t, actualExpiration.Sub(testSecret.Expiration().Time()), time.Second)
			})

			t.Run("it should set TTL on access limit", func(t *testing.T) {
				val, err := redisClient.TTL(ctx, keys.AccessLimit()).Result()
				actualExpiration := time.Now().Add(val).UTC()
				require.NoError(t, err)
				assert.LessOrEqual(t, actualExpiration.Sub(testSecret.Expiration().Time()), time.Second)
			})

			if tc.filename != "" {
				t.Run("it should store filename in redis", func(t *testing.T) {
					val, err := redisClient.Get(ctx, keys.Filename()).Result()
					require.NoError(t, err)
					assert.Equal(t, testSecret.Filename, val)
				})

				t.Run("it should set TTL on filename", func(t *testing.T) {
					val, err := redisClient.TTL(ctx, keys.Filename()).Result()
					actualExpiration := time.Now().Add(val).UTC()
					require.NoError(t, err)
					assert.LessOrEqual(t, actualExpiration.Sub(testSecret.Expiration().Time()), time.Second)
				})
			} else {
				t.Run("it should not store filename in redis", func(t *testing.T) {
					val, err := redisClient.Exists(ctx, keys.Filename()).Result()
					require.NoError(t, err)
					assert.Equal(t, int64(0), val)
				})
			}
		})
	}
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

	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	var testCases = []struct {
		name          string
		filename      string
		setupOldStyle bool
	}{
		{name: "has filename", filename: "retrieved-doc.txt", setupOldStyle: false},
		{name: "does not have filename", filename: "", setupOldStyle: true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("and secret %s", tc.name), func(t *testing.T) {
			var secret models.Secret
			var keys *redis.RedisKey

			if tc.setupOldStyle {
				// Test backward compatibility: manually create old-style secret without filename key
				id := testhelpers.RandomId(t)
				keys = redis.NewRedisKeySet(id)

				_ = redisClient.Set(ctx, keys.ContentType(), models.ContentTypeFile, time.Minute).Err()
				_ = redisClient.Set(ctx, keys.Content(), expected.CipherText, time.Minute).Err()
				_ = redisClient.Set(ctx, keys.AccessLimit(), expected.AccessLimit, time.Minute).Err()
				_ = redisClient.Set(ctx, keys.Access(), 0, time.Minute).Err()
				_ = redisClient.Set(ctx, keys.ExpirationEpoch(), expected.ExpirationEpoch, time.Minute).Err()
				// Intentionally NOT setting filename key

				secret = expected
				secret.ID = id
			} else {
				secret = expected
				secret.ID = testhelpers.RandomId(t)
				secret.Filename = tc.filename
				keys = redis.NewRedisKeySet(secret.ID)
				require.NoError(t, sut.WriteSecret(ctx, secret))
			}

			t.Cleanup(func() {
				_ = redisClient.Del(ctx, keys.AllKeys()...).Err()
			})

			actual := sut.ReadSecret(ctx, secret.ID)

			t.Run("it should return ID", func(t *testing.T) {
				assert.Equal(t, secret.ID, actual.ID)
			})

			t.Run("it should return content", func(t *testing.T) {
				assert.Equal(t, secret.CipherText, actual.CipherText)
			})

			t.Run("it should return access count", func(t *testing.T) {
				assert.Equal(t, 0, actual.AccessCount)
			})

			t.Run("it should return access limit", func(t *testing.T) {
				assert.Equal(t, secret.AccessLimit, actual.AccessLimit)
			})

			if tc.filename != "" {
				t.Run("it should return filename", func(t *testing.T) {
					assert.Equal(t, tc.filename, actual.Filename)
				})
			} else {
				t.Run("it should return empty filename", func(t *testing.T) {
					assert.Equal(t, "", actual.Filename)
				})
			}
		})
	}
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
