package settings

import (
	"github.com/spf13/viper"
	"testing"
)

func TestAppConfiguration(t *testing.T) {
	t.Run("when MaxFileSizeMB is not set", func(t *testing.T) {
		viper.Reset()
		app := NewAppConfiguration()

		t.Run("it should return default value of 8 MB", func(t *testing.T) {
			result := app.MaxFileSizeMB()
			expected := 8

			if result != expected {
				t.Errorf("expected MaxFileSizeMB to be %d, got %d", expected, result)
			}
		})
	})

	t.Run("when MaxFileSizeMB is set via environment", func(t *testing.T) {
		viper.Reset()
		viper.Set("app.maxFileSizeMB", 16)
		app := NewAppConfiguration()

		t.Run("it should return the configured value", func(t *testing.T) {
			result := app.MaxFileSizeMB()
			expected := 16

			if result != expected {
				t.Errorf("expected MaxFileSizeMB to be %d, got %d", expected, result)
			}
		})
	})
}
