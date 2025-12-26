//go:build integration
// +build integration

package vault

import (
	"cellar/pkg/cryptography/vault"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestWhenGettingHealth(t *testing.T) {
	ctx := context.Background()
	cfg := settings.NewConfiguration()
	sut, err := vault.NewEncryptionClient(ctx, cfg.Encryption().Vault())
	if err != nil {
		t.Error(err)
	}
	actual := sut.Health(ctx)

	t.Run("should return name", testhelpers.EqualsF("vault", strings.ToLower(actual.Name)))
	t.Run("should return healthy status", testhelpers.EqualsF("healthy", strings.ToLower(actual.Status)))
	t.Run("should return version", testhelpers.NotEqualsF("", actual.Version))
}

func TestVaultEncryption(t *testing.T) {
	ctx := context.Background()
	cfg := settings.NewConfiguration()
	sut, err := vault.NewEncryptionClient(ctx, cfg.Encryption().Vault())
	if err != nil {
		t.Error(err)
	}
	content := "some secret content"
	encrypted, err := sut.Encrypt(ctx, []byte(content))
	t.Run("when encrypting", func(t *testing.T) {
		t.Run("should not return error", testhelpers.EqualsF(nil, err))
		t.Run("should return encrypted in the right format", testhelpers.AssertF(func() bool {
			matched, err := regexp.MatchString("^vault:v\\d+:\\S+$", encrypted)
			if err != nil {
				t.Error(err)
			}
			return matched
		}(), fmt.Sprintf("expected vault encrypted text, but was '%s'", encrypted)))
	})
	decrypted, err := sut.Decrypt(ctx, encrypted)
	t.Run("when decrypting", func(t *testing.T) {
		t.Run("should not return error", testhelpers.EqualsF(nil, err))
		t.Run("should return initial content", testhelpers.EqualsF(content, string(decrypted)))
	})
}
