package commands_test

import (
	"cellar/pkg/commands"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"github.com/golang/mock/gomock"
	"strings"
	"testing"
)

func TestGetHealth(t *testing.T) {
	ctrl := gomock.NewController(t)

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

	health := commands.GetHealth(dataStore, encryption)

	t.Run("status should be Healthy", testhelpers.EqualsF("healthy", strings.ToLower(health.Status)))
	t.Run("should return host", testhelpers.NotEqualsF("", health.Host))
	t.Run("should return datastore health", testhelpers.EqualsF(dataStoreHealth, health.Datastore))
	t.Run("should return encryption health", testhelpers.EqualsF(encryptionHealth, health.Encryption))
}
