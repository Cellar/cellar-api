//go:build acceptance
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

func TestWhenAccessingSecretContent(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV2(t, cfg, models.ContentTypeText, content, 10)

	path := fmt.Sprintf("%s/v2/secrets/%s/access", cfg.App().ClientAddress(), secret.ID)
	resp, err := http.Post(path, "application/json", nil)
	testhelpers.OkF(err)

	t.Run("status should be ok", testhelpers.EqualsF(http.StatusOK, resp.StatusCode))

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretContentResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should match", testhelpers.EqualsF(secret.ID, actual.ID))
	t.Run("content should match", testhelpers.EqualsF(content, actual.Content))
}

func TestWhenAccessingSecretFile(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV2(t, cfg, models.ContentTypeFile, content, 10)

	path := fmt.Sprintf("%s/v2/secrets/%s/access", cfg.App().ClientAddress(), secret.ID)
	resp, err := http.Post(path, "application/json", nil)
	testhelpers.OkF(err)

	t.Run("status should be ok", testhelpers.EqualsF(http.StatusOK, resp.StatusCode))

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	t.Run("file should match", testhelpers.EqualsF(content, string(responseBody)))
	t.Run("Content-Type should be octet-stream", testhelpers.EqualsF("application/octet-stream", resp.Header.Get("Content-Type")))
}

func TestWhenAccessingSecretContentForSecretThatDoesntExist(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	path := fmt.Sprintf("%s/v2/secrets/%s/content", cfg.App().ClientAddress(), testhelpers.RandomId(t))
	resp, err := http.Get(path)
	testhelpers.OkF(err)

	t.Run("status should be not found", testhelpers.EqualsF(http.StatusNotFound, resp.StatusCode))
}

func TestWhenAccessingSecretWithAccessLimitOfOne(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	content := "Super Secret Test Content"
	secret := testhelpers.CreateSecretV2(t, cfg, models.ContentTypeText, content, 1)

	path := fmt.Sprintf("%s/v2/secrets/%s/access", cfg.App().ClientAddress(), secret.ID)
	response1, err := http.Post(path, "application/json", nil)
	testhelpers.OkF(err)
	response2, err := http.Post(path, "application/json", nil)
	testhelpers.OkF(err)

	t.Run("first request status should be ok", testhelpers.EqualsF(http.StatusOK, response1.StatusCode))
	t.Run("second request status should be not found", testhelpers.EqualsF(http.StatusNotFound, response2.StatusCode))

	defer response1.Body.Close()

	responseBody, err := ioutil.ReadAll(response1.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretContentResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("response1 id should match", testhelpers.EqualsF(secret.ID, actual.ID))
	t.Run("response1 content should match", testhelpers.EqualsF(content, actual.Content))
}
