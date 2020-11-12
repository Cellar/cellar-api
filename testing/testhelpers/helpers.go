package testhelpers

import (
	"bytes"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func GetConfiguration() settings.IConfiguration {
	cfg := settings.NewConfiguration()
	return cfg
}

func RandomId(tb testing.TB) string {
	bytes := make([]byte, 25)
	_, err := rand.Read(bytes)
	Ok(tb, err)

	return hex.EncodeToString(bytes)
}

func GetRedisClient(cfg settings.Configuration) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis().Host(), cfg.Redis().Port()),
		Password: cfg.Redis().Password(),
		DB:       cfg.Redis().DB(),
	})
}

func CreateSecret(t *testing.T, cfg settings.IConfiguration, content string, accessLimit int) models.SecretMetadataResponse {
	secret := map[string]interface{}{
		"content":          content,
		"access_limit":     accessLimit,
		"expiration_epoch": EpochFromNow(time.Hour),
	}
	body, err := json.Marshal(secret)
	Ok(t, err)

	createResp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	OkF(err)

	defer func() {
		Ok(t, createResp.Body.Close())
	}()

	responseBody, err := ioutil.ReadAll(createResp.Body)
	Ok(t, err)

	var createdSecret models.SecretMetadataResponse
	Ok(t, json.Unmarshal(responseBody, &createdSecret))

	return createdSecret
}

func EpochFromNow(duration time.Duration) int64 {
	return time.Now().UTC().Add(duration).Unix()
}
