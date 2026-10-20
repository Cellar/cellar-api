package testhelpers

import (
	"cellar/pkg/datastore"
	pkgredis "cellar/pkg/datastore/redis"
	pkgvalkey "cellar/pkg/datastore/valkey"
	"cellar/pkg/ratelimit"
	"cellar/pkg/settings"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type Backend struct {
	Name        string
	Store       datastore.DataStore
	WindowStore ratelimit.WindowStore
	Raw         *redis.Client
}

func EnabledDatastoreName(cfg settings.IConfiguration) string {
	if cfg.Datastore().Valkey().Enabled() {
		return "valkey"
	}
	return "redis"
}

func ForEachBackend(t *testing.T, fn func(t *testing.T, backend Backend)) {
	cfg := settings.NewConfiguration()
	redisCfg := cfg.Datastore().Redis()
	valkeyCfg := cfg.Datastore().Valkey()

	if !redisCfg.Enabled() && !valkeyCfg.Enabled() {
		t.Fatal("no datastore enabled: set DATASTORE_REDIS_ENABLED and/or DATASTORE_VALKEY_ENABLED")
	}

	if redisCfg.Enabled() {
		t.Run("given redis", func(t *testing.T) {
			store := pkgredis.NewDataStore(redisCfg)
			raw := GetRedisClient(redisCfg)
			t.Cleanup(func() {
				_ = store.Close()
				_ = raw.Close()
			})
			fn(t, Backend{Name: "redis", Store: store, WindowStore: ratelimit.NewRedisWindowStore(store.Client()), Raw: raw})
		})
	}

	if valkeyCfg.Enabled() {
		t.Run("given valkey", func(t *testing.T) {
			store, err := pkgvalkey.NewDataStore(valkeyCfg)
			require.NoError(t, err)
			raw := redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", valkeyCfg.Host(), valkeyCfg.Port()),
				Username: valkeyCfg.Username(),
				Password: valkeyCfg.Password(),
				DB:       valkeyCfg.DB(),
			})
			t.Cleanup(func() {
				_ = store.Close()
				_ = raw.Close()
			})
			fn(t, Backend{Name: "valkey", Store: store, WindowStore: ratelimit.NewValkeyWindowStore(store.Client()), Raw: raw})
		})
	}
}
