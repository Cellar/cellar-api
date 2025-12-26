//go:build acceptance
// +build acceptance

package secrets

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenDeletingASecret(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	client := &http.Client{}
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV2(t, cfg, models.ContentTypeText, content, 10)

	path := fmt.Sprintf("%s/v2/secrets/%s", cfg.App().ClientAddress(), secret.ID)
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	resp, err := client.Do(req)
	require.NoError(t, err)

	t.Run("it should return no content status", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("it should delete the secret", func(t *testing.T) {
		resp, err := http.Get(path)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestWhenDeletingSecretThatDoesntExist(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	client := &http.Client{}
	path := fmt.Sprintf("%s/v2/secrets/%s", cfg.App().ClientAddress(), testhelpers.RandomId(t))
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	t.Run("it should return not found status", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
