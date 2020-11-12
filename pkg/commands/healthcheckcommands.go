package commands

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/models"
	"os"
)

func GetHealth(dataStore datastore.DataStore, encryption cryptography.Encryption) models.HealthResponse {
	name, err := os.Hostname()
	if err != nil {
		name = "Unknown"
	}

	return *models.NewHealthResponse(name, dataStore.Health(), encryption.Health())
}
