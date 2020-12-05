package commands

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"os"
)

func GetHealth(appConfig settings.IAppConfiguration, dataStore datastore.DataStore, encryption cryptography.Encryption) models.HealthResponse {
	name, err := os.Hostname()
	if err != nil {
		name = "Unknown"
	}

	return *models.NewHealthResponse(name, appConfig.Version(), dataStore.Health(), encryption.Health())
}
