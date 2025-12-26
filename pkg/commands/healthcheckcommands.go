package commands

import (
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"context"
	"os"
)

// GetHealth checks the health status of the application and its dependencies.
// Returns a health response containing the application version, hostname, and health status of datastore and encryption services.
// The context can be used to cancel the health checks before completion.
func GetHealth(ctx context.Context, appConfig settings.IAppConfiguration, dataStore datastore.DataStore, encryption cryptography.Encryption) models.HealthResponse {
	name, err := os.Hostname()
	if err != nil {
		name = "Unknown"
	}

	return *models.NewHealthResponse(name, appConfig.Version(), dataStore.Health(ctx), encryption.Health(ctx))
}
