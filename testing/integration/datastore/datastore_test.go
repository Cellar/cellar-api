//go:build integration
// +build integration

package datastore

import (
	"cellar/pkg/datastore/keyvalue"
	"cellar/pkg/models"
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

func newSecret(t *testing.T, lifetime time.Duration) models.Secret {
	return models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      testhelpers.RandomId(t),
		ContentType:     models.ContentTypeText,
		AccessLimit:     50,
		ExpirationEpoch: testhelpers.EpochFromNow(lifetime),
	}
}

func TestWhenGettingHealth(t *testing.T) {
	testhelpers.ForEachBackend(t, func(t *testing.T, backend testhelpers.Backend) {
		actual := backend.Store.Health(context.Background())

		t.Run("it should return the backend name", func(t *testing.T) {
			assert.True(t, strings.EqualFold(backend.Name, actual.Name), "expected name %q (case insensitive), got %q", backend.Name, actual.Name)
		})

		t.Run("it should return healthy status", func(t *testing.T) {
			assert.True(t, strings.EqualFold("healthy", actual.Status), "expected status to be 'healthy' (case insensitive), got %s", actual.Status)
		})

		t.Run("it should return a known version", func(t *testing.T) {
			assert.NotEqual(t, "Unknown", actual.Version)
		})
	})
}

func TestWhenWritingSecret(t *testing.T) {
	testhelpers.ForEachBackend(t, func(t *testing.T, backend testhelpers.Backend) {
		ctx := context.Background()
		raw := backend.Raw

		var testCases = []struct {
			name     string
			filename string
		}{
			{name: "has filename", filename: "test-file.pdf"},
			{name: "does not have filename", filename: ""},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("and secret %s", tc.name), func(t *testing.T) {
				testSecret := newSecret(t, time.Minute)
				testSecret.Filename = tc.filename

				keys := keyvalue.NewKeySet(testSecret.ID)
				err := backend.Store.WriteSecret(ctx, testSecret)

				t.Cleanup(func() {
					_ = raw.Del(ctx, keys.AllKeys()...).Err()
				})

				t.Run("it should not return error", func(t *testing.T) {
					assert.NoError(t, err)
				})

				stored := []struct {
					name     string
					key      string
					expected string
				}{
					{"content type", keys.ContentType(), testSecret.ContentType},
					{"cipher text", keys.Content(), testSecret.CipherText},
					{"access limit", keys.AccessLimit(), strconv.Itoa(testSecret.AccessLimit)},
					{"access count", keys.Access(), strconv.Itoa(testSecret.AccessCount)},
					{"expiration", keys.ExpirationEpoch(), strconv.FormatInt(testSecret.ExpirationEpoch, 10)},
				}
				if tc.filename != "" {
					stored = append(stored, struct {
						name     string
						key      string
						expected string
					}{"filename", keys.Filename(), testSecret.Filename})
				}

				for _, s := range stored {
					t.Run("it should store the "+s.name, func(t *testing.T) {
						val, err := raw.Get(ctx, s.key).Result()
						require.NoError(t, err)
						assert.Equal(t, s.expected, val)
					})

					t.Run("it should set a TTL matching the expiration on the "+s.name, func(t *testing.T) {
						val, err := raw.TTL(ctx, s.key).Result()
						require.NoError(t, err)
						actualExpiration := time.Now().Add(val).UTC()
						assert.LessOrEqual(t, actualExpiration.Sub(testSecret.Expiration().Time()).Abs(), 2*time.Second)
					})
				}

				if tc.filename == "" {
					t.Run("it should not store a filename", func(t *testing.T) {
						val, err := raw.Exists(ctx, keys.Filename()).Result()
						require.NoError(t, err)
						assert.Equal(t, int64(0), val)
					})
				}
			})
		}
	})
}

func TestWhenSecretExpires(t *testing.T) {
	testhelpers.ForEachBackend(t, func(t *testing.T, backend testhelpers.Backend) {
		ctx := context.Background()
		secret := newSecret(t, 2*time.Second)
		secret.Filename = "expiring.txt"
		keys := keyvalue.NewKeySet(secret.ID)

		require.NoError(t, backend.Store.WriteSecret(ctx, secret))
		t.Cleanup(func() {
			_ = backend.Raw.Del(ctx, keys.AllKeys()...).Err()
		})

		t.Run("it should be readable before it expires", func(t *testing.T) {
			assert.NotNil(t, backend.Store.ReadSecret(ctx, secret.ID))
		})

		t.Run("and the expiration passes", func(t *testing.T) {
			require.Eventually(t, func() bool {
				return backend.Store.ReadSecret(ctx, secret.ID) == nil
			}, 10*time.Second, 100*time.Millisecond)

			for _, key := range keys.AllKeys() {
				t.Run("it should remove "+key[strings.LastIndex(key, ":")+1:]+" from the datastore", func(t *testing.T) {
					require.Eventually(t, func() bool {
						exists, err := backend.Raw.Exists(ctx, key).Result()
						return err == nil && exists == 0
					}, 5*time.Second, 100*time.Millisecond)
				})
			}
		})
	})
}

func TestWhenReadingSecret(t *testing.T) {
	testhelpers.ForEachBackend(t, func(t *testing.T, backend testhelpers.Backend) {
		ctx := context.Background()
		raw := backend.Raw
		expected := newSecret(t, time.Minute)

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
				var keys *keyvalue.KeySet

				if tc.setupOldStyle {
					id := testhelpers.RandomId(t)
					keys = keyvalue.NewKeySet(id)

					_ = raw.Set(ctx, keys.ContentType(), models.ContentTypeFile, time.Minute).Err()
					_ = raw.Set(ctx, keys.Content(), expected.CipherText, time.Minute).Err()
					_ = raw.Set(ctx, keys.AccessLimit(), expected.AccessLimit, time.Minute).Err()
					_ = raw.Set(ctx, keys.Access(), 0, time.Minute).Err()
					_ = raw.Set(ctx, keys.ExpirationEpoch(), expected.ExpirationEpoch, time.Minute).Err()

					secret = expected
					secret.ID = id
				} else {
					secret = expected
					secret.ID = testhelpers.RandomId(t)
					secret.Filename = tc.filename
					keys = keyvalue.NewKeySet(secret.ID)
					require.NoError(t, backend.Store.WriteSecret(ctx, secret))
				}

				t.Cleanup(func() {
					_ = raw.Del(ctx, keys.AllKeys()...).Err()
				})

				actual := backend.Store.ReadSecret(ctx, secret.ID)
				require.NotNil(t, actual)

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

				t.Run("it should return the filename", func(t *testing.T) {
					assert.Equal(t, tc.filename, actual.Filename)
				})
			})
		}
	})
}

func TestWhenDeletingSecret(t *testing.T) {
	testhelpers.ForEachBackend(t, func(t *testing.T, backend testhelpers.Backend) {
		ctx := context.Background()
		secret := newSecret(t, time.Minute)
		secret.Filename = "deleted.txt"
		keys := keyvalue.NewKeySet(secret.ID)

		require.NoError(t, backend.Store.WriteSecret(ctx, secret))
		t.Cleanup(func() {
			_ = backend.Raw.Del(ctx, keys.AllKeys()...).Err()
		})

		deleted, err := backend.Store.DeleteSecret(ctx, secret.ID)

		t.Run("it should return true", func(t *testing.T) {
			assert.True(t, deleted)
		})

		t.Run("it should not return error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		for _, key := range keys.AllKeys() {
			t.Run("it should not find "+key[strings.LastIndex(key, ":")+1:], func(t *testing.T) {
				val, err := backend.Raw.Exists(ctx, key).Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), val)
			})
		}
	})
}

func TestWhenIncreasingSecretAccess(t *testing.T) {
	testhelpers.ForEachBackend(t, func(t *testing.T, backend testhelpers.Backend) {
		ctx := context.Background()
		secret := newSecret(t, time.Minute)
		keys := keyvalue.NewKeySet(secret.ID)

		require.NoError(t, backend.Store.WriteSecret(ctx, secret))
		t.Cleanup(func() {
			_ = backend.Raw.Del(ctx, keys.AllKeys()...).Err()
		})

		actual, err := backend.Store.IncreaseAccessCount(ctx, secret.ID)

		t.Run("it should not return error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("it should increase access count", func(t *testing.T) {
			assert.Equal(t, int64(1), actual)
		})

		t.Run("it should increase access count in datastore", func(t *testing.T) {
			val, err := backend.Raw.Get(ctx, keys.Access()).Int()
			require.NoError(t, err)
			assert.Equal(t, 1, val)
		})

		t.Run("it should not change the access count TTL", func(t *testing.T) {
			val, err := backend.Raw.TTL(ctx, keys.Access()).Result()
			require.NoError(t, err)
			assert.Greater(t, val, 50*time.Second)
		})

		t.Run("it should not increase access limit in datastore", func(t *testing.T) {
			val, err := backend.Raw.Get(ctx, keys.AccessLimit()).Int()
			require.NoError(t, err)
			assert.Equal(t, secret.AccessLimit, val)
		})
	})
}
