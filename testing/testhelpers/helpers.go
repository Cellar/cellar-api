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
	"io"
	"io/ioutil"
	"mime/multipart"
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

func GetRedisClient(cfg settings.IRedisConfiguration) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host(), cfg.Port()),
		Password: cfg.Password(),
		DB:       cfg.DB(),
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

func PostFormData(t *testing.T, uri string, formData map[string]string, fileFormData map[string]string) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)


	for key, content := range fileFormData {
		filename := RandomId(t)
		fileContents := bytes.NewBufferString(content)
		part, err := writer.CreateFormFile(key, filename)
		Ok(t, err)

		_, err = io.Copy(part, fileContents)
		Ok(t, err)
	}

	for key, val := range formData {
		Ok(t, writer.WriteField(key, val))
	}

	err := writer.Close()
	Ok(t, err)

	request, err := http.NewRequest("POST", uri, body)
	Ok(t, err)

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	Ok(t, err)

	return resp
}
