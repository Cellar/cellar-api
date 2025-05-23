package commands_test

import (
	"cellar/pkg/commands"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
)

func TestGetHealth(t *testing.T) {
	ctrl := gomock.NewController(t)
	cfg := settings.NewConfiguration()

	encryption := mocks.NewMockEncryption(ctrl)
	encryptionHealth := *models.NewHealth(
		"test_encryption",
		models.Healthy,
		"1.0.0",
	)
	encryption.EXPECT().
		Health().
		Return(encryptionHealth)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStoreHealth := *models.NewHealth(
		"test_datastore",
		models.Healthy,
		"0.1.0",
	)
	dataStore.EXPECT().
		Health().
		Return(dataStoreHealth)

	health := commands.GetHealth(cfg.App(), dataStore, encryption)

	t.Run("status should be Healthy", testhelpers.EqualsF("healthy", strings.ToLower(health.Status)))
	t.Run("should return host", testhelpers.NotEqualsF("", health.Host))
	t.Run("should return version", testhelpers.EqualsF(cfg.App().Version(), health.Version))
	t.Run("should return datastore health", testhelpers.EqualsF(dataStoreHealth, health.Datastore))
	t.Run("should return encryption health", testhelpers.EqualsF(encryptionHealth, health.Encryption))
}
