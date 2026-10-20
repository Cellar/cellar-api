package datastore

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRedisConfiguration(t *testing.T) {
	t.Run("when nothing is configured", func(t *testing.T) {
		viper.Reset()
		sut := NewDatastoreConfiguration().Redis()

		t.Run("it should be disabled", func(t *testing.T) {
			assert.False(t, sut.Enabled())
		})
	})

	t.Run("when enabled is set", func(t *testing.T) {
		viper.Reset()
		viper.Set("datastore.redis.enabled", true)
		sut := NewDatastoreConfiguration().Redis()

		t.Run("it should be enabled", func(t *testing.T) {
			assert.True(t, sut.Enabled())
		})
	})
}

func TestValkeyConfiguration(t *testing.T) {
	t.Run("when nothing is configured", func(t *testing.T) {
		viper.Reset()
		sut := NewDatastoreConfiguration().Valkey()

		defaults := []struct {
			name     string
			actual   any
			expected any
		}{
			{name: "be disabled", actual: sut.Enabled(), expected: false},
			{name: "default host to localhost", actual: sut.Host(), expected: "localhost"},
			{name: "default port to 6379", actual: sut.Port(), expected: 6379},
			{name: "default username to empty", actual: sut.Username(), expected: ""},
			{name: "default password to empty", actual: sut.Password(), expected: ""},
			{name: "default db to 0", actual: sut.DB(), expected: 0},
			{name: "default tls to disabled", actual: sut.TLSEnabled(), expected: false},
		}

		for _, tc := range defaults {
			t.Run("it should "+tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.actual)
			})
		}
	})

	t.Run("when everything is configured", func(t *testing.T) {
		viper.Reset()
		viper.Set("datastore.valkey.enabled", true)
		viper.Set("datastore.valkey.host", "valkey.example.com")
		viper.Set("datastore.valkey.port", 6380)
		viper.Set("datastore.valkey.username", "cellar")
		viper.Set("datastore.valkey.password", "hunter2")
		viper.Set("datastore.valkey.db", 2)
		viper.Set("datastore.valkey.tls_enabled", true)
		sut := NewDatastoreConfiguration().Valkey()

		configured := []struct {
			name     string
			actual   any
			expected any
		}{
			{name: "be enabled", actual: sut.Enabled(), expected: true},
			{name: "return the host", actual: sut.Host(), expected: "valkey.example.com"},
			{name: "return the port", actual: sut.Port(), expected: 6380},
			{name: "return the username", actual: sut.Username(), expected: "cellar"},
			{name: "return the password", actual: sut.Password(), expected: "hunter2"},
			{name: "return the db", actual: sut.DB(), expected: 2},
			{name: "enable tls", actual: sut.TLSEnabled(), expected: true},
		}

		for _, tc := range configured {
			t.Run("it should "+tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.actual)
			})
		}
	})
}
