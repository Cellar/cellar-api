package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

type HealthStatus int

const (
	Healthy = iota
	Degraded
	Unhealthy
)

func (s HealthStatus) String() string {
	return [...]string{"Healthy", "Degraded", "Unhealthy"}[s]
}

func (s HealthStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *HealthStatus) UnmarshalJSON(data []byte) error {
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}
	if stat, ok := map[string]HealthStatus{
		"healthy":   Healthy,
		"degraded":  Degraded,
		"unhealthy": Unhealthy,
	}[strings.ToLower(dataStr)]; !ok {
		return fmt.Errorf("unknown status: '%s'", dataStr)
	} else {
		*s = stat
	}

	return nil
}

func assessHealth(statuses []HealthStatus) HealthStatus {
	if len(statuses) == 1 {
		return statuses[0]
	}

	allUnhealthy := true
	for _, status := range statuses {
		allUnhealthy = allUnhealthy && status == Unhealthy
		if !allUnhealthy {
			if status != Healthy {
				return Degraded
			}
		}
	}

	if allUnhealthy {
		return Unhealthy
	}

	return Healthy
}

type HealthResponse struct {
	Host       string `json:"host" example:"localhost"`
	Status     string `json:"status" example:"healthy"`
	Datastore  Health `json:"datastore"`
	Encryption Health `json:"encryption"`
}

type Health struct {
	Name       string       `json:"name" example:"Redis"`
	statusEnum HealthStatus `json:"-"`
	Status     string       `json:"status" example:"healthy"`
	Version    string       `json:"version" example:"1.0.0"`
}

func NewHealth(name string, status HealthStatus, version string) *Health {
	return &Health{
		Name:       name,
		Status:     status.String(),
		statusEnum: status,
		Version:    version,
	}
}

func NewHealthResponse(host string, dataStoreHealth, encryptionHealth Health) *HealthResponse {
	status := assessHealth([]HealthStatus{dataStoreHealth.statusEnum, encryptionHealth.statusEnum})

	return &HealthResponse{
		Host:       host,
		Status:     status.String(),
		Datastore:  dataStoreHealth,
		Encryption: encryptionHealth,
	}
}
