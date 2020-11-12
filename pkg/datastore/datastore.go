package datastore

import (
	"cellar/pkg/models"
)

var Key = "DATASTORE"

//go:generate mockgen -destination=../mocks/mock_datastore.go -package=mocks . DataStore
type DataStore interface {
	Health() models.Health
	WriteSecret(secret models.Secret) (err error)
	ReadSecret(id string) (secret *models.Secret)
	IncreaseAccessCount(id string) (accessCount int64, err error)
	DeleteSecret(id string) (found bool, err error)
}
