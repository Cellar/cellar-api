// +build acceptance

package secrets

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestWhenCreatingASecretFromContent(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()
	expectedAccessLimit := 100

	body := map[string]string{
		"content": "Super Secret Test Content",
		"access_limit": strconv.Itoa(expectedAccessLimit),
		"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
	}

	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, nil)

	t.Run("status is created", testhelpers.EqualsF(http.StatusCreated, resp.StatusCode))

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretMetadataResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should not be empty", testhelpers.NotEqualsF("", actual.ID))
	t.Run("access limit should set", testhelpers.EqualsF(expectedAccessLimit, actual.AccessLimit))
	t.Run("expiration should be set", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format()))
}

func TestWhenCreatingASecretFromFile(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()
	expectedAccessLimit := 100

	body := map[string]string{
		"access_limit": strconv.Itoa(expectedAccessLimit),
		"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
	}

	fileContent := map[string]string{ "file": "Super Secret Test Content" }
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, fileContent)

	t.Run("status is created", testhelpers.EqualsF(http.StatusCreated, resp.StatusCode))

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretMetadataResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should not be empty", testhelpers.NotEqualsF("", actual.ID))
	t.Run("access limit should set", testhelpers.EqualsF(expectedAccessLimit, actual.AccessLimit))
	t.Run("expiration should be set", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format()))
}

func TestWhenCreatingASecretAndExpirationIsTooShort(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Minute * 9).UTC()

	body := map[string]string{
		"content": "Super Secret Test Content",
		"access_limit": "100",
		"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
	}
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, nil)

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}

func TestWhenCreatingASecretAndExpirationIsInThePast(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour * -24).UTC()

	body := map[string]string{
		"content": "Super Secret Test Content",
		"access_limit": "100",
		"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
	}
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, nil)

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}

func TestWhenCreatingASecretFromContentWithoutAccessLimit(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	body := map[string]string{
		"content": "Super Secret Test Content",
		"expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10),
	}
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, nil)

	t.Run("status is created", testhelpers.EqualsF(http.StatusCreated, resp.StatusCode))

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretMetadataResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should not be empty", testhelpers.NotEqualsF("", actual.ID))
	t.Run("access limit should set to 0", testhelpers.EqualsF(0, actual.AccessLimit))
	t.Run("expiration should be set", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format()))
}

func TestWhenCreatingASecretFromFileWithoutAccessLimit(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	body := map[string]string{ "expiration_epoch": strconv.FormatInt(expectedExpiration.Unix(), 10) }
	fileContent := map[string]string{ "file": "Super Secret Test Content" }
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, fileContent)

	t.Run("status is created", testhelpers.EqualsF(http.StatusCreated, resp.StatusCode))

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretMetadataResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should not be empty", testhelpers.NotEqualsF("", actual.ID))
	t.Run("access limit should set to 0", testhelpers.EqualsF(0, actual.AccessLimit))
	t.Run("expiration should be set", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format()))
}

func TestWhenCreatingASecretWithoutExpiration(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	body := map[string]string{
		"content": "Super Secret Test Content",
		"access_limit": "100",
	}
	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, nil)

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}

func TestWhenCreatingASecretWithoutContentOrFile(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	body := map[string]string{
		"access_limit": "100",
		"duration": strconv.FormatInt(time.Now().Add(time.Hour).UTC().Unix(), 10),
	}

	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, nil)

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}

func TestWhenCreatingASecretWithContentAndFile(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	body := map[string]string{
		"content": "Super Secret Test Content",
		"access_limit": "100",
		"duration": strconv.FormatInt(time.Now().Add(time.Hour).UTC().Unix(), 10),
	}
	fileContent := map[string]string{ "file": "Super Secret Test Content" }

	resp := testhelpers.PostFormData(t, cfg.App().ClientAddress() + "/v2/secrets", body, fileContent)

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}
