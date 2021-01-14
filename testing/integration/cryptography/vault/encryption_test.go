// +build integration

package vault

import (
	"cellar/pkg/cryptography/vault"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestWhenGettingHealth(t *testing.T) {
	cfg := settings.NewConfiguration()
	sut, err := vault.NewEncryptionClient(cfg.Vault())
	if err != nil {
		t.Error(err)
	}
	actual := sut.Health()

	t.Run("should return name", testhelpers.EqualsF("vault", strings.ToLower(actual.Name)))
	t.Run("should return healthy status", testhelpers.EqualsF("healthy", strings.ToLower(actual.Status)))
	t.Run("should return version", testhelpers.NotEqualsF("", actual.Version))
}

func TestEncryption(t *testing.T) {
	cfg := settings.NewConfiguration()
	sut, err := vault.NewEncryptionClient(cfg.Vault())
	if err != nil {
		t.Error(err)
	}
	content := "some secret content"
	encrypted, err := sut.Encrypt([]byte(content))
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
	decrypted, err := sut.Decrypt(encrypted)
	t.Run("when decrypting", func(t *testing.T) {
		t.Run("should not return error", testhelpers.EqualsF(nil, err))
		t.Run("should return initial content", testhelpers.EqualsF(content, decrypted))
	})
}
