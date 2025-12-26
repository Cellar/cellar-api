//go:build integration
// +build integration

package vault

import (
	"cellar/pkg/cryptography/vault"
	"cellar/pkg/settings"
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenGettingHealth(t *testing.T) {
	ctx := context.Background()
	cfg := settings.NewConfiguration()
	sut, err := vault.NewEncryptionClient(ctx, cfg.Encryption().Vault())
	require.NoError(t, err)

	actual := sut.Health(ctx)

	t.Run("it should return vault name", func(t *testing.T) {
		assert.Equal(t, "vault", actual.Name)
	})

	t.Run("it should return healthy status", func(t *testing.T) {
		assert.Equal(t, "healthy", actual.Status)
	})

	t.Run("it should return version", func(t *testing.T) {
		assert.NotEmpty(t, actual.Version)
	})
}

func TestVaultEncryption(t *testing.T) {
	ctx := context.Background()
	cfg := settings.NewConfiguration()
	sut, err := vault.NewEncryptionClient(ctx, cfg.Encryption().Vault())
	require.NoError(t, err)

	content := "some secret content"
	encrypted, err := sut.Encrypt(ctx, []byte(content))

	t.Run("when encrypting", func(t *testing.T) {
		t.Run("it should not return error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("it should return encrypted in the right format", func(t *testing.T) {
			matched, matchErr := regexp.MatchString("^vault:v\\d+:\\S+$", encrypted)
			require.NoError(t, matchErr)
			assert.True(t, matched)
		})
	})

	decrypted, decryptErr := sut.Decrypt(ctx, encrypted)

	t.Run("when decrypting", func(t *testing.T) {
		t.Run("it should not return error", func(t *testing.T) {
			assert.NoError(t, decryptErr)
		})

		t.Run("it should return initial content", func(t *testing.T) {
			assert.Equal(t, content, string(decrypted))
		})
	})
}
