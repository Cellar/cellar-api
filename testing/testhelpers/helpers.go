package testhelpers

import (
	"bytes"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"cellar/pkg/settings/datastore"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func GetConfiguration() settings.IConfiguration {
	cfg := settings.NewConfiguration()
	return cfg
}

func RandomId(tb testing.TB) string {
	bytes := make([]byte, 25)
	_, err := rand.Read(bytes)
	require.NoError(tb, err)

	return hex.EncodeToString(bytes)
}

func GetRedisClient(cfg datastore.IRedisConfiguration) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host(), cfg.Port()),
		Password: cfg.Password(),
		DB:       cfg.DB(),
	})
}
func CreateSecretV1(t *testing.T, cfg settings.IConfiguration, content string, accessLimit int) models.SecretMetadataResponse {
	secret := map[string]interface{}{
		"access_limit":     accessLimit,
		"expiration_epoch": EpochFromNow(time.Hour),
		"content":          content,
	}

	body, err := json.Marshal(secret)
	require.NoError(t, err)

	createResp, err := http.Post(cfg.App().ClientAddress()+"/v1/secrets", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, createResp.Body.Close())
	}()

	responseBody, err := io.ReadAll(createResp.Body)
	require.NoError(t, err)

	var createdSecret models.SecretMetadataResponse
	require.NoError(t, json.Unmarshal(responseBody, &createdSecret))

	return createdSecret
}

func CreateSecretV2(t *testing.T, cfg settings.IConfiguration, contentType models.ContentType, content string, accessLimit int) models.SecretMetadataResponseV2 {
	formData := map[string]string{
		"access_limit":     strconv.Itoa(accessLimit),
		"expiration_epoch": strconv.FormatInt(EpochFromNow(time.Hour), 10),
	}
	fileFormData := map[string]string{}
	if contentType == models.ContentTypeText {
		formData["content"] = content
	} else {
		fileFormData["file"] = content
	}
	createResp := PostFormData(t, cfg.App().ClientAddress()+"/v2/secrets", formData, fileFormData)
	defer func() {
		require.NoError(t, createResp.Body.Close())
	}()

	responseBody, err := io.ReadAll(createResp.Body)
	require.NoError(t, err)

	var createdSecret models.SecretMetadataResponseV2
	require.NoError(t, json.Unmarshal(responseBody, &createdSecret))

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
		require.NoError(t, err)

		_, err = io.Copy(part, fileContents)
		require.NoError(t, err)
	}

	for key, val := range formData {
		require.NoError(t, writer.WriteField(key, val))
	}

	err := writer.Close()
	require.NoError(t, err)

	request, err := http.NewRequest("POST", uri, body)
	require.NoError(t, err)

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	require.NoError(t, err)

	return resp
}
