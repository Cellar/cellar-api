// +build acceptance

package healthcheck

import (
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	cfg := testhelpers.GetConfiguration()
	resp, err := http.Get(cfg.App().ClientAddress() + "/health-check")
	testhelpers.OkF(err)

	defer func() {
		testhelpers.Ok(t, resp.Body.Close())
	}()

	t.Run("status is ok", testhelpers.EqualsF(200, resp.StatusCode))

	responseBody, err := ioutil.ReadAll(resp.Body)
	testhelpers.Ok(t, err)

	var health models.HealthResponse
	testhelpers.Ok(t, json.Unmarshal(responseBody, &health))

	t.Run("status should be Healthy", testhelpers.EqualsF("healthy", strings.ToLower(health.Status)))
	t.Run("should return host", testhelpers.NotEqualsF("", health.Host))
	t.Run("should return non empty version", testhelpers.NotEqualsF("", health.Version))

	t.Run("should return datastore name", testhelpers.EqualsF("redis", strings.ToLower(health.Datastore.Name)))
	t.Run("should return datastore healthy status", testhelpers.EqualsF("healthy", strings.ToLower(health.Datastore.Status)))
	t.Run("should return datastore version", testhelpers.NotEqualsF("", health.Datastore.Version))

	t.Run("should return encryption name", testhelpers.EqualsF("vault", strings.ToLower(health.Encryption.Name)))
	t.Run("should return encryption healthy status", testhelpers.EqualsF("healthy", strings.ToLower(health.Encryption.Status)))
	t.Run("should return encryption version", testhelpers.NotEqualsF("", health.Encryption.Version))
}
