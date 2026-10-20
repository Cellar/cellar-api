package keyvalue

import (
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/models"
	"context"
	"errors"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const idFieldKey = "datastore_key"

var ErrExpired = errors.New("secret expiration must be in the future")

type DataStore struct {
	client Client
	server Server
	logger *log.Entry
}

func NewDataStore(client Client, server Server) *DataStore {
	return &DataStore{
		client: client,
		server: server,
		logger: log.WithFields(log.Fields{
			"context":  "datastore",
			"instance": strings.ToLower(server.Name),
		}),
	}
}

func (store DataStore) Health(ctx context.Context) models.Health {
	status := models.HealthStatus(models.Unhealthy)
	version := "Unknown"

	if err := pkgerrors.CheckContext(ctx); err != nil {
		return *models.NewHealth(store.server.Name, status, version)
	}

	if info, err := store.client.Info(ctx, "server"); err == nil {
		prefix := store.server.VersionKey + ":"
		for _, line := range strings.Split(strings.ReplaceAll(info, "\r\n", "\n"), "\n") {
			if strings.HasPrefix(line, prefix) {
				status = models.Healthy
				version = strings.TrimSpace(strings.TrimPrefix(line, prefix))
				break
			}
		}
	}

	return *models.NewHealth(store.server.Name, status, version)
}

func (store DataStore) WriteSecret(ctx context.Context, secret models.Secret) error {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return err
	}

	ttl := secret.Duration()
	if ttl <= 0 {
		return ErrExpired
	}

	keySet := NewKeySet(secret.ID)
	store.logger.WithField(idFieldKey, keySet.id).Debug("writing secret to datastore")

	values := []struct {
		key   string
		value string
	}{
		{keySet.Access(), "0"},
		{keySet.AccessLimit(), strconv.Itoa(secret.AccessLimit)},
		{keySet.ContentType(), secret.ContentType},
		{keySet.Content(), secret.CipherText},
		{keySet.ExpirationEpoch(), strconv.FormatInt(secret.ExpirationEpoch, 10)},
	}
	if secret.Filename != "" {
		values = append(values, struct {
			key   string
			value string
		}{keySet.Filename(), secret.Filename})
	}

	for _, v := range values {
		if err := store.client.Set(ctx, v.key, v.value, ttl); err != nil {
			return err
		}
	}

	return nil
}

func (store DataStore) ReadSecret(ctx context.Context, id string) *models.Secret {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil
	}

	keySet := NewKeySet(id)
	store.logger.WithField(idFieldKey, keySet.id).Debug("reading secret from datastore")

	accessLimit, err := store.getInt(ctx, keySet.AccessLimit())
	if err != nil {
		return nil
	}

	contentType, err := store.client.Get(ctx, keySet.ContentType())
	if err != nil {
		return nil
	}

	content, err := store.client.Get(ctx, keySet.Content())
	if err != nil {
		return nil
	}

	accessCount, err := store.getInt(ctx, keySet.Access())
	if err != nil {
		return nil
	}

	expirationEpoch, err := store.getInt64(ctx, keySet.ExpirationEpoch())
	if err != nil {
		return nil
	}

	filename, err := store.client.Get(ctx, keySet.Filename())
	if err != nil {
		filename = ""
	}

	return &models.Secret{
		ID:              id,
		CipherText:      content,
		ContentType:     contentType,
		Filename:        filename,
		AccessCount:     accessCount,
		AccessLimit:     accessLimit,
		ExpirationEpoch: expirationEpoch,
	}
}

func (store DataStore) IncreaseAccessCount(ctx context.Context, id string) (int64, error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return 0, err
	}

	keySet := NewKeySet(id)
	store.logger.WithField(idFieldKey, keySet.id).Debug("increasing secret access count in datastore")
	return store.client.Incr(ctx, keySet.Access())
}

func (store DataStore) DeleteSecret(ctx context.Context, id string) (bool, error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return false, err
	}

	keySet := NewKeySet(id)
	store.logger.WithField(idFieldKey, keySet.id).Debug("deleting secret from datastore")
	numDeleted, err := store.client.Del(ctx, keySet.AllKeys()...)
	return numDeleted > 0, err
}

func (store DataStore) Close() error {
	return store.client.Close()
}

func (store DataStore) getInt(ctx context.Context, key string) (int, error) {
	value, err := store.client.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func (store DataStore) getInt64(ctx context.Context, key string) (int64, error) {
	value, err := store.client.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, 64)
}
