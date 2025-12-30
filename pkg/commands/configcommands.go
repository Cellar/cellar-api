package commands

import (
	"cellar/pkg/models"
	"cellar/pkg/settings"
)

// GetConfig returns the current runtime configuration limits as a domain model.
// This is a pure function that transforms settings into a limits model.
// The controller layer wraps this in a ConfigResponse for HTTP responses.
func GetConfig(cfg settings.IConfiguration) models.LimitsConfig {
	return models.LimitsConfig{
		MaxFileSizeMB:        cfg.App().MaxFileSizeMB(),
		MaxAccessCount:       cfg.App().MaxAccessCount(),
		MaxExpirationSeconds: cfg.App().MaxExpirationSeconds(),
	}
}
