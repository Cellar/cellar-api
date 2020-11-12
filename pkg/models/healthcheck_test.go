package models_test

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"testing"
)

func NewHealthResponseTest(t *testing.T, dataStoreStatus, encryptionStatus, expectedStatus models.HealthStatus) {
	host := "host"
	dataStoreHealth := *models.NewHealth(
		"dataStore",
		dataStoreStatus,
		"1.1.1",
	)
	encryptionHealth := *models.NewHealth(
		"encryption",
		encryptionStatus,
		"2.2.2",
	)
	expected := models.HealthResponse{
		Host:       host,
		Status:     expectedStatus.String(),
		Datastore:  dataStoreHealth,
		Encryption: encryptionHealth,
	}
	actual := *models.NewHealthResponse(
		"host",
		dataStoreHealth,
		encryptionHealth,
	)
	testhelpers.Equals(t, expected, actual)
}

func TestAssessHealth_WhenAllHealthy(t *testing.T) {
	NewHealthResponseTest(t, models.Healthy, models.Healthy, models.Healthy)
}

func TestAssessHealth_WhenAllUnhealthy(t *testing.T) {
	NewHealthResponseTest(t, models.Unhealthy, models.Unhealthy, models.Unhealthy)
}

func TestAssessHealth_WhenAllDegraded(t *testing.T) {
	NewHealthResponseTest(t, models.Degraded, models.Degraded, models.Degraded)
}

func TestAssessHealth_WhenOneUnhealthy(t *testing.T) {
	NewHealthResponseTest(t, models.Healthy, models.Unhealthy, models.Degraded)
}

func TestAssessHealth_WhenOneDegraded(t *testing.T) {
	NewHealthResponseTest(t, models.Healthy, models.Degraded, models.Degraded)
}
