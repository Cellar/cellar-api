//go:build acceptance
// +build acceptance

package secrets

import (
	"bytes"
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestWhenCreatingASecret(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	expected := struct {
		Content         string `json:"content"`
		AccessLimit     int    `json:"access_limit"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		AccessLimit:     100,
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	testhelpers.Ok(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	testhelpers.OkF(err)

	defer resp.Body.Close()

	t.Run("status is created", testhelpers.EqualsF(http.StatusCreated, resp.StatusCode))

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretMetadataResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should not be empty", testhelpers.NotEqualsF("", actual.ID))
	t.Run("access limit should set", testhelpers.EqualsF(expected.AccessLimit, actual.AccessLimit))
	t.Run("expiration should be set", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format()))
}

func TestWhenCreatingASecretAndExpirationIsTooShort(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Minute * 9).UTC()

	expected := struct {
		Content         string `json:"content"`
		AccessLimit     int    `json:"access_limit"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		AccessLimit:     100,
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	testhelpers.Ok(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	testhelpers.OkF(err)

	defer resp.Body.Close()

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}

func TestWhenCreatingASecretAndExpirationIsInThePast(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour * -24).UTC()

	expected := struct {
		Content         string `json:"content"`
		AccessLimit     int    `json:"access_limit"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		AccessLimit:     100,
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	testhelpers.Ok(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	testhelpers.OkF(err)

	defer resp.Body.Close()

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}

func TestWhenCreatingASecretWithoutAccessLimit(t *testing.T) {
	cfg := testhelpers.GetConfiguration()

	expectedExpiration := time.Now().Add(time.Hour).UTC()

	expected := struct {
		Content         string `json:"content"`
		ExpirationEpoch int64  `json:"expiration_epoch"`
	}{
		Content:         "Super Secret Test Content",
		ExpirationEpoch: expectedExpiration.Unix(),
	}
	body, err := json.Marshal(expected)
	testhelpers.Ok(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	testhelpers.OkF(err)

	defer resp.Body.Close()

	t.Run("status is created", testhelpers.EqualsF(http.StatusCreated, resp.StatusCode))

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var actual models.SecretMetadataResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &actual))

	t.Run("id should not be empty", testhelpers.NotEqualsF("", actual.ID))
	t.Run("access limit should set to 0", testhelpers.EqualsF(0, actual.AccessLimit))
	t.Run("expiration should be set", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), actual.Expiration.Format()))
}

func TestWhenCreatingASecretWithoutDuration(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	expected := map[string]interface{}{
		"content":      "Super Secret Test Content",
		"access_limit": 100,
	}
	body, err := json.Marshal(expected)
	testhelpers.Ok(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	testhelpers.OkF(err)

	defer resp.Body.Close()

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}

func TestWhenCreatingASecretWithoutContent(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	expected := map[string]interface{}{
		"access_limit": 100,
		"duration":     60,
	}
	body, err := json.Marshal(expected)
	testhelpers.Ok(t, err)

	resp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	testhelpers.OkF(err)

	defer resp.Body.Close()

	t.Run("status is bad request", testhelpers.EqualsF(http.StatusBadRequest, resp.StatusCode))
}
