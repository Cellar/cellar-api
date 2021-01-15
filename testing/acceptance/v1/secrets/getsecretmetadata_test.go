// +build acceptance

package secrets

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestWhenGettingSecret(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV1(t, cfg, content, 10)

	path := fmt.Sprintf("%s/v1/secrets/%s", cfg.App().ClientAddress(), secret.ID)
	resp, err := http.Get(path)
	testhelpers.OkF(err)

	t.Run("status should be ok", testhelpers.EqualsF(http.StatusOK, resp.StatusCode))

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretMetadataResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should match", testhelpers.EqualsF(secret.ID, actual.ID))
	t.Run("access count should match", testhelpers.EqualsF(secret.AccessCount, actual.AccessCount))
	t.Run("access limit count should match", testhelpers.EqualsF(secret.AccessLimit, actual.AccessLimit))
	t.Run("expiration should match", testhelpers.EqualsF(secret.Expiration.Format(), actual.Expiration.Format()))
}

func TestWhenGettingSecretThatDoesntExist(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	path := fmt.Sprintf("%s/v1/secrets/%s", cfg.App().ClientAddress(), testhelpers.RandomId(t))
	resp, err := http.Get(path)
	testhelpers.OkF(err)

	t.Run("status should be not found", testhelpers.EqualsF(http.StatusNotFound, resp.StatusCode))
}
