// +build acceptance

package secrets

import (
	"cellar/testing/testhelpers"
	"fmt"
	"net/http"
	"testing"
)

func TestWhenDeletingASecret(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	client := &http.Client{}
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecret(t, cfg, content, 10)

	path := fmt.Sprintf("%s/v2/secrets/%s", cfg.App().ClientAddress(), secret.ID)
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	resp, err := client.Do(req)
	testhelpers.OkF(err)

	t.Run("status should be no content", testhelpers.EqualsF(http.StatusNoContent, resp.StatusCode))

	t.Run("secret should be deleted", func(t *testing.T) {
		resp, err := http.Get(path)
		testhelpers.Ok(t, err)

		testhelpers.Equals(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestWhenDeletingSecretThatDoesntExist(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	client := &http.Client{}
	path := fmt.Sprintf("%s/v2/secrets/%s", cfg.App().ClientAddress(), testhelpers.RandomId(t))
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	testhelpers.OkF(err)

	resp, err := client.Do(req)
	testhelpers.OkF(err)

	t.Run("status should be not found", testhelpers.EqualsF(http.StatusNotFound, resp.StatusCode))
}
