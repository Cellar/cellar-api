package settings

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestAppConfiguration(t *testing.T) {
	t.Run("when testing MaxFileSizeMB", func(t *testing.T) {
		testCases := []struct {
			name          string
			setValue      *int
			expectedValue int
			reason        string
		}{
			{
				name:          "not set",
				setValue:      nil,
				expectedValue: 8,
				reason:        "default value of 8 MB",
			},
			{
				name:          "set to valid value",
				setValue:      intPtr(16),
				expectedValue: 16,
				reason:        "configured value",
			},
			{
				name:          "set to zero",
				setValue:      intPtr(0),
				expectedValue: 0,
				reason:        "0 indicates file uploads are disabled",
			},
			{
				name:          "set to negative value",
				setValue:      intPtr(-5),
				expectedValue: 0,
				reason:        "0 as minimum value",
			},
		}

		for _, tc := range testCases {
			t.Run("and "+tc.name, func(t *testing.T) {
				viper.Reset()
				if tc.setValue != nil {
					viper.Set("app.max_file_size_mb", *tc.setValue)
				}
				app := NewAppConfiguration()

				t.Run("it should return "+tc.reason, func(t *testing.T) {
					result := app.MaxFileSizeMB()
					assert.Equal(t, tc.expectedValue, result)
				})
			})
		}
	})

	t.Run("when testing MaxAccessCount", func(t *testing.T) {
		testCases := []struct {
			name          string
			setValue      *int
			expectedValue int
			reason        string
		}{
			{
				name:          "not set",
				setValue:      nil,
				expectedValue: 100,
				reason:        "default value of 100",
			},
			{
				name:          "set to valid value",
				setValue:      intPtr(50),
				expectedValue: 50,
				reason:        "configured value",
			},
			{
				name:          "set to zero",
				setValue:      intPtr(0),
				expectedValue: 1,
				reason:        "1 as minimum value",
			},
			{
				name:          "set to negative value",
				setValue:      intPtr(-10),
				expectedValue: 1,
				reason:        "1 as minimum value",
			},
		}

		for _, tc := range testCases {
			t.Run("and "+tc.name, func(t *testing.T) {
				viper.Reset()
				if tc.setValue != nil {
					viper.Set("app.max_access_count", *tc.setValue)
				}
				app := NewAppConfiguration()

				t.Run("it should return "+tc.reason, func(t *testing.T) {
					result := app.MaxAccessCount()
					assert.Equal(t, tc.expectedValue, result)
				})
			})
		}
	})

	t.Run("when testing MaxExpirationSeconds", func(t *testing.T) {
		testCases := []struct {
			name          string
			setValue      *int
			expectedValue int
			reason        string
		}{
			{
				name:          "not set",
				setValue:      nil,
				expectedValue: 604800,
				reason:        "default value of 604800 seconds",
			},
			{
				name:          "set to valid value",
				setValue:      intPtr(3600),
				expectedValue: 3600,
				reason:        "configured value",
			},
			{
				name:          "set below minimum of 900",
				setValue:      intPtr(600),
				expectedValue: 900,
				reason:        "900 as minimum value (15 minutes)",
			},
			{
				name:          "set to negative value",
				setValue:      intPtr(-300),
				expectedValue: 900,
				reason:        "900 as minimum value (15 minutes)",
			},
		}

		for _, tc := range testCases {
			t.Run("and "+tc.name, func(t *testing.T) {
				viper.Reset()
				if tc.setValue != nil {
					viper.Set("app.max_expiration_seconds", *tc.setValue)
				}
				app := NewAppConfiguration()

				t.Run("it should return "+tc.reason, func(t *testing.T) {
					result := app.MaxExpirationSeconds()
					assert.Equal(t, tc.expectedValue, result)
				})
			})
		}
	})
}

func intPtr(i int) *int {
	return &i
}
