package commands_test

import (
	"cellar/pkg/commands"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetHealth(t *testing.T) {
	t.Run("when getting health status", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		cfg := settings.NewConfiguration()

		encryption := mocks.NewMockEncryption(ctrl)
		encryptionHealth := *models.NewHealth("test_encryption", models.Healthy, "1.0.0")
		encryption.EXPECT().Health(gomock.Any()).Return(encryptionHealth)

		dataStore := mocks.NewMockDataStore(ctrl)
		dataStoreHealth := *models.NewHealth("test_datastore", models.Healthy, "0.1.0")
		dataStore.EXPECT().Health(gomock.Any()).Return(dataStoreHealth)

		health := commands.GetHealth(context.Background(), cfg.App(), dataStore, encryption)

		t.Run("it should have healthy status", func(t *testing.T) {
			assert.Equal(t, "Healthy", health.Status)
		})

		t.Run("it should return host", func(t *testing.T) {
			assert.NotEmpty(t, health.Host)
		})

		t.Run("it should return version", func(t *testing.T) {
			assert.Equal(t, cfg.App().Version(), health.Version)
		})

		t.Run("it should return datastore health", func(t *testing.T) {
			assert.Equal(t, dataStoreHealth, health.Datastore)
		})

		t.Run("it should return encryption health", func(t *testing.T) {
			assert.Equal(t, encryptionHealth, health.Encryption)
		})
	})
}
