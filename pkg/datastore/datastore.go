package datastore

import (
	"cellar/pkg/models"
	"context"
)

var Key = "DATASTORE"

//go:generate mockgen -destination=../mocks/mock_datastore.go -package=mocks . DataStore
type DataStore interface {
	Health(ctx context.Context) models.Health
	WriteSecret(ctx context.Context, secret models.Secret) (err error)
	ReadSecret(ctx context.Context, id string) (secret *models.Secret)
	IncreaseAccessCount(ctx context.Context, id string) (accessCount int64, err error)
	DeleteSecret(ctx context.Context, id string) (found bool, err error)
}
