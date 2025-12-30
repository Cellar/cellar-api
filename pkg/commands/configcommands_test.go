package commands

import (
	"cellar/pkg/settings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	t.Run("when getting config", func(t *testing.T) {
		cfg := settings.NewConfiguration()

		result := GetConfig(cfg)

		t.Run("it should return limits", func(t *testing.T) {
			assert.NotNil(t, result)
		})

		t.Run("it should return maxFileSizeMB from configuration", func(t *testing.T) {
			assert.Equal(t, cfg.App().MaxFileSizeMB(), result.MaxFileSizeMB)
		})

		t.Run("it should return maxAccessCount from configuration", func(t *testing.T) {
			assert.Equal(t, cfg.App().MaxAccessCount(), result.MaxAccessCount)
		})

		t.Run("it should return maxExpirationSeconds from configuration", func(t *testing.T) {
			assert.Equal(t, cfg.App().MaxExpirationSeconds(), result.MaxExpirationSeconds)
		})
	})
}
