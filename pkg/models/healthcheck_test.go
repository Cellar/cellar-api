package models_test

import (
	"cellar/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhenCreatingHealthResponse(t *testing.T) {
	t.Run("when all components are healthy", func(t *testing.T) {
		host := "host"
		appVersion := "1.0.0"
		dataStoreHealth := *models.NewHealth("dataStore", models.Healthy, "1.1.1")
		encryptionHealth := *models.NewHealth("encryption", models.Healthy, "2.2.2")

		actual := *models.NewHealthResponse(host, appVersion, dataStoreHealth, encryptionHealth)

		t.Run("it should set host", func(t *testing.T) {
			assert.Equal(t, host, actual.Host)
		})

		t.Run("it should set version", func(t *testing.T) {
			assert.Equal(t, appVersion, actual.Version)
		})

		t.Run("it should set overall status to healthy", func(t *testing.T) {
			assert.Equal(t, "Healthy", actual.Status)
		})

		t.Run("it should set datastore health", func(t *testing.T) {
			assert.Equal(t, dataStoreHealth, actual.Datastore)
		})

		t.Run("it should set encryption health", func(t *testing.T) {
			assert.Equal(t, encryptionHealth, actual.Encryption)
		})
	})

	t.Run("when all components are unhealthy", func(t *testing.T) {
		host := "host"
		appVersion := "1.0.0"
		dataStoreHealth := *models.NewHealth("dataStore", models.Unhealthy, "1.1.1")
		encryptionHealth := *models.NewHealth("encryption", models.Unhealthy, "2.2.2")

		actual := *models.NewHealthResponse(host, appVersion, dataStoreHealth, encryptionHealth)

		t.Run("it should set overall status to unhealthy", func(t *testing.T) {
			assert.Equal(t, "Unhealthy", actual.Status)
		})
	})

	t.Run("when all components are degraded", func(t *testing.T) {
		host := "host"
		appVersion := "1.0.0"
		dataStoreHealth := *models.NewHealth("dataStore", models.Degraded, "1.1.1")
		encryptionHealth := *models.NewHealth("encryption", models.Degraded, "2.2.2")

		actual := *models.NewHealthResponse(host, appVersion, dataStoreHealth, encryptionHealth)

		t.Run("it should set overall status to degraded", func(t *testing.T) {
			assert.Equal(t, "Degraded", actual.Status)
		})
	})

	t.Run("when one component is unhealthy", func(t *testing.T) {
		host := "host"
		appVersion := "1.0.0"
		dataStoreHealth := *models.NewHealth("dataStore", models.Healthy, "1.1.1")
		encryptionHealth := *models.NewHealth("encryption", models.Unhealthy, "2.2.2")

		actual := *models.NewHealthResponse(host, appVersion, dataStoreHealth, encryptionHealth)

		t.Run("it should set overall status to degraded", func(t *testing.T) {
			assert.Equal(t, "Degraded", actual.Status)
		})
	})

	t.Run("when one component is degraded", func(t *testing.T) {
		host := "host"
		appVersion := "1.0.0"
		dataStoreHealth := *models.NewHealth("dataStore", models.Healthy, "1.1.1")
		encryptionHealth := *models.NewHealth("encryption", models.Degraded, "2.2.2")

		actual := *models.NewHealthResponse(host, appVersion, dataStoreHealth, encryptionHealth)

		t.Run("it should set overall status to degraded", func(t *testing.T) {
			assert.Equal(t, "Degraded", actual.Status)
		})
	})
}
