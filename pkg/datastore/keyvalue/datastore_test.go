package keyvalue_test

import (
	"cellar/pkg/datastore/keyvalue"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	redisServer  = keyvalue.Server{Name: "Redis", VersionKey: "redis_version"}
	valkeyServer = keyvalue.Server{Name: "Valkey", VersionKey: "valkey_version"}
	errBoom      = errors.New("boom")
)

func newSut(t *testing.T, server keyvalue.Server) (*keyvalue.DataStore, *mocks.MockClient) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	return keyvalue.NewDataStore(client, server), client
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func ttlNear(expected time.Duration) gomock.Matcher {
	return gomock.Cond(func(actual time.Duration) bool {
		diff := expected - actual
		return diff >= 0 && diff < 5*time.Second
	})
}

func TestDataStoreHealth(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		description     string
		server          keyvalue.Server
		info            string
		infoErr         error
		expectedName    string
		expectedStatus  string
		expectedVersion string
	}{
		{
			description:     "a redis server",
			server:          redisServer,
			info:            "# Server\r\nredis_version:8.2.1\r\nredis_mode:standalone\r\n",
			expectedName:    "Redis",
			expectedStatus:  "Healthy",
			expectedVersion: "8.2.1",
		},
		{
			description:     "a valkey server",
			server:          valkeyServer,
			info:            "# Server\r\nredis_version:7.2.4\r\nserver_name:valkey\r\nvalkey_version:9.0.0\r\n",
			expectedName:    "Valkey",
			expectedStatus:  "Healthy",
			expectedVersion: "9.0.0",
		},
		{
			description:     "a valkey server that does not report a valkey version",
			server:          valkeyServer,
			info:            "# Server\r\nredis_version:7.2.4\r\n",
			expectedName:    "Valkey",
			expectedStatus:  "Unhealthy",
			expectedVersion: "Unknown",
		},
		{
			description:     "an unreachable server",
			server:          redisServer,
			infoErr:         errBoom,
			expectedName:    "Redis",
			expectedStatus:  "Unhealthy",
			expectedVersion: "Unknown",
		},
	}

	for _, tc := range testCases {
		t.Run("when checking health of "+tc.description, func(t *testing.T) {
			sut, client := newSut(t, tc.server)
			client.EXPECT().Info(ctx, "server").Return(tc.info, tc.infoErr)
			actual := sut.Health(ctx)

			t.Run("it should report the server name", func(t *testing.T) {
				assert.Equal(t, tc.expectedName, actual.Name)
			})

			t.Run("it should report the status", func(t *testing.T) {
				assert.Equal(t, tc.expectedStatus, actual.Status)
			})

			t.Run("it should report the version", func(t *testing.T) {
				assert.Equal(t, tc.expectedVersion, actual.Version)
			})
		})
	}

	t.Run("when the context is cancelled", func(t *testing.T) {
		sut, _ := newSut(t, redisServer)
		actual := sut.Health(cancelledContext())

		t.Run("it should report unhealthy", func(t *testing.T) {
			assert.Equal(t, "Unhealthy", actual.Status)
		})
	})
}

func TestDataStoreWriteSecret(t *testing.T) {
	ctx := context.Background()
	expiration := time.Now().Add(time.Hour)
	secret := models.Secret{
		ID:              "abc",
		CipherText:      "cipher",
		ContentType:     models.ContentTypeText,
		AccessLimit:     5,
		ExpirationEpoch: expiration.Unix(),
	}
	keys := keyvalue.NewKeySet(secret.ID)
	ttl := ttlNear(secret.Duration())

	expectBaseKeys := func(client *mocks.MockClient) {
		client.EXPECT().Set(ctx, keys.Access(), "0", ttl).Return(nil)
		client.EXPECT().Set(ctx, keys.AccessLimit(), "5", ttl).Return(nil)
		client.EXPECT().Set(ctx, keys.ContentType(), models.ContentTypeText, ttl).Return(nil)
		client.EXPECT().Set(ctx, keys.Content(), "cipher", ttl).Return(nil)
		client.EXPECT().Set(ctx, keys.ExpirationEpoch(), strconv.FormatInt(secret.ExpirationEpoch, 10), ttl).Return(nil)
	}

	t.Run("when writing a secret without a filename", func(t *testing.T) {
		sut, client := newSut(t, redisServer)
		expectBaseKeys(client)

		t.Run("it should write every key with the secret ttl", func(t *testing.T) {
			assert.NoError(t, sut.WriteSecret(ctx, secret))
		})
	})

	t.Run("when writing a secret with a filename", func(t *testing.T) {
		sut, client := newSut(t, redisServer)
		expectBaseKeys(client)
		client.EXPECT().Set(ctx, keys.Filename(), "notes.txt", ttl).Return(nil)
		withFile := secret
		withFile.Filename = "notes.txt"

		t.Run("it should also write the filename with the secret ttl", func(t *testing.T) {
			assert.NoError(t, sut.WriteSecret(ctx, withFile))
		})
	})

	t.Run("when the client fails to write", func(t *testing.T) {
		sut, client := newSut(t, redisServer)
		client.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errBoom)

		t.Run("it should return the error", func(t *testing.T) {
			assert.ErrorIs(t, sut.WriteSecret(ctx, secret), errBoom)
		})
	})

	t.Run("when the secret has already expired", func(t *testing.T) {
		sut, _ := newSut(t, redisServer)
		expired := secret
		expired.ExpirationEpoch = time.Now().Add(-time.Minute).Unix()

		t.Run("it should refuse to write a secret without a ttl", func(t *testing.T) {
			assert.ErrorIs(t, sut.WriteSecret(ctx, expired), keyvalue.ErrExpired)
		})
	})

	t.Run("when the context is cancelled", func(t *testing.T) {
		sut, _ := newSut(t, redisServer)

		t.Run("it should return an error", func(t *testing.T) {
			assert.Error(t, sut.WriteSecret(cancelledContext(), secret))
		})
	})
}

func TestDataStoreReadSecret(t *testing.T) {
	ctx := context.Background()
	keys := keyvalue.NewKeySet("abc")

	stored := map[string]string{
		keys.AccessLimit():     "5",
		keys.ContentType():     models.ContentTypeFile,
		keys.Content():         "cipher",
		keys.Access():          "2",
		keys.ExpirationEpoch(): "1700000000",
		keys.Filename():        "notes.txt",
	}

	expectReads := func(client *mocks.MockClient, missing string) {
		client.EXPECT().Get(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, key string) (string, error) {
			value, ok := stored[key]
			if !ok || key == missing {
				return "", errBoom
			}
			return value, nil
		}).AnyTimes()
	}

	t.Run("when every key exists", func(t *testing.T) {
		sut, client := newSut(t, redisServer)
		expectReads(client, "")
		actual := sut.ReadSecret(ctx, "abc")

		t.Run("it should return the secret", func(t *testing.T) {
			assert.Equal(t, &models.Secret{
				ID:              "abc",
				CipherText:      "cipher",
				ContentType:     models.ContentTypeFile,
				Filename:        "notes.txt",
				AccessCount:     2,
				AccessLimit:     5,
				ExpirationEpoch: 1700000000,
			}, actual)
		})
	})

	t.Run("when the filename is missing", func(t *testing.T) {
		sut, client := newSut(t, redisServer)
		expectReads(client, keys.Filename())
		actual := sut.ReadSecret(ctx, "abc")

		t.Run("it should return an empty filename", func(t *testing.T) {
			assert.Empty(t, actual.Filename)
		})
	})

	for _, required := range []string{keys.AccessLimit(), keys.ContentType(), keys.Content(), keys.Access(), keys.ExpirationEpoch()} {
		t.Run("when "+required+" is missing", func(t *testing.T) {
			sut, client := newSut(t, redisServer)
			expectReads(client, required)

			t.Run("it should return nil", func(t *testing.T) {
				assert.Nil(t, sut.ReadSecret(ctx, "abc"))
			})
		})
	}

	t.Run("when the context is cancelled", func(t *testing.T) {
		sut, _ := newSut(t, redisServer)

		t.Run("it should return nil", func(t *testing.T) {
			assert.Nil(t, sut.ReadSecret(cancelledContext(), "abc"))
		})
	})
}

func TestDataStoreIncreaseAccessCount(t *testing.T) {
	ctx := context.Background()

	t.Run("when increasing the access count", func(t *testing.T) {
		sut, client := newSut(t, redisServer)
		client.EXPECT().Incr(ctx, keyvalue.NewKeySet("abc").Access()).Return(int64(3), nil)
		actual, _ := sut.IncreaseAccessCount(ctx, "abc")

		t.Run("it should return the new count", func(t *testing.T) {
			assert.Equal(t, int64(3), actual)
		})
	})

	t.Run("when the context is cancelled", func(t *testing.T) {
		sut, _ := newSut(t, redisServer)
		_, err := sut.IncreaseAccessCount(cancelledContext(), "abc")

		t.Run("it should return an error", func(t *testing.T) {
			assert.Error(t, err)
		})
	})
}

func TestDataStoreDeleteSecret(t *testing.T) {
	ctx := context.Background()
	allKeys := keyvalue.NewKeySet("abc").AllKeys()

	testCases := []struct {
		description string
		deleted     int64
		expected    bool
	}{
		{description: "keys exist", deleted: 6, expected: true},
		{description: "no keys exist", deleted: 0, expected: false},
	}

	for _, tc := range testCases {
		t.Run("when deleting a secret and "+tc.description, func(t *testing.T) {
			sut, client := newSut(t, redisServer)
			client.EXPECT().Del(ctx, allKeys).Return(tc.deleted, nil)
			actual, _ := sut.DeleteSecret(ctx, "abc")

			t.Run("it should report whether the secret was found", func(t *testing.T) {
				assert.Equal(t, tc.expected, actual)
			})
		})
	}

	t.Run("when the context is cancelled", func(t *testing.T) {
		sut, _ := newSut(t, redisServer)
		_, err := sut.DeleteSecret(cancelledContext(), "abc")

		t.Run("it should return an error", func(t *testing.T) {
			assert.Error(t, err)
		})
	})
}

func TestDataStoreClose(t *testing.T) {
	t.Run("when closing the datastore", func(t *testing.T) {
		sut, client := newSut(t, redisServer)
		client.EXPECT().Close().Return(errBoom)

		t.Run("it should close the client", func(t *testing.T) {
			assert.ErrorIs(t, sut.Close(), errBoom)
		})
	})
}
