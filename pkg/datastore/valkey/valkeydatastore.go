package valkey

import (
	"cellar/pkg/datastore/keyvalue"
	"cellar/pkg/settings/datastore"
	"context"
	"crypto/tls"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/valkey-io/valkey-go"
)

var Server = keyvalue.Server{Name: "Valkey", VersionKey: "valkey_version"}

type DataStore struct {
	*keyvalue.DataStore
	client valkey.Client
}

func NewDataStore(configuration datastore.IValkeyConfiguration) (*DataStore, error) {
	address := fmt.Sprintf("%s:%d", configuration.Host(), configuration.Port())
	logger := log.WithFields(log.Fields{"context": "datastore", "instance": "valkey", "address": address})
	logger.Debug("initializing valkey configuration")
	if configuration.Password() == "" {
		logger.Warn("valkey password is empty")
	}

	client, err := valkey.NewClient(ClientOption(configuration))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to valkey at %s: %w", address, err)
	}

	return &DataStore{
		DataStore: keyvalue.NewDataStore(clientAdapter{client}, Server),
		client:    client,
	}, nil
}

func ClientOption(configuration datastore.IValkeyConfiguration) valkey.ClientOption {
	option := valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", configuration.Host(), configuration.Port())},
		Username:    configuration.Username(),
		Password:    configuration.Password(),
		SelectDB:    configuration.DB(),
	}
	if configuration.TLSEnabled() {
		option.TLSConfig = &tls.Config{
			ServerName: configuration.Host(),
			MinVersion: tls.VersionTLS12,
		}
	}
	return option
}

func (store *DataStore) Client() valkey.Client {
	return store.client
}

type clientAdapter struct {
	client valkey.Client
}

func (c clientAdapter) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Do(ctx, c.client.B().Set().Key(key).Value(value).Px(ttl).Build()).Error()
}

func (c clientAdapter) Get(ctx context.Context, key string) (string, error) {
	return c.client.Do(ctx, c.client.B().Get().Key(key).Build()).ToString()
}

func (c clientAdapter) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Do(ctx, c.client.B().Incr().Key(key).Build()).AsInt64()
}

func (c clientAdapter) Del(ctx context.Context, keys ...string) (int64, error) {
	cmds := make(valkey.Commands, len(keys))
	for i, key := range keys {
		cmds[i] = c.client.B().Del().Key(key).Build()
	}

	var deleted int64
	for _, result := range c.client.DoMulti(ctx, cmds...) {
		count, err := result.AsInt64()
		if err != nil {
			return 0, err
		}
		deleted += count
	}
	return deleted, nil
}

func (c clientAdapter) Info(ctx context.Context, section string) (string, error) {
	return c.client.Do(ctx, c.client.B().Info().Section(section).Build()).ToString()
}

func (c clientAdapter) Close() error {
	c.client.Close()
	return nil
}
