package v2

import (
	"bytes"
	"cellar/pkg/cryptography"
	"cellar/pkg/datastore"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/gin-gonic/gin"
)

var Equals = testhelpers.Equals

func TestCreateSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("when calling the create endpoint", func(t *testing.T) {
		var router *gin.Engine
		var cfg settings.IConfiguration
		var ctrl *gomock.Controller
		var mockDataStore *mocks.MockDataStore
		var mockEncryption *mocks.MockEncryption

		createMultipartRequest := func(fileContent []byte, filename string) *http.Request {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, _ := writer.CreateFormFile("file", filename)
			part.Write(fileContent)
			writer.WriteField("expiration_epoch", "9999999999")
			writer.Close()

			req, _ := http.NewRequest("POST", "/v2/secrets", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			return req
		}

		setupRouter := func() {
			router = gin.New()
			cfg = settings.NewConfiguration()
			ctrl = gomock.NewController(t)
			mockDataStore = mocks.NewMockDataStore(ctrl)
			mockEncryption = mocks.NewMockEncryption(ctrl)

			router.Use(func(c *gin.Context) {
				c.Set(settings.Key, cfg)
				c.Set(datastore.Key, mockDataStore)
				c.Set(cryptography.Key, mockEncryption)
				c.Next()
			})

			router.POST("/v2/secrets", CreateSecret)
		}

		t.Run("and file size exceeds limit", func(t *testing.T) {
			setupRouter()
			maxSizeMB := cfg.App().MaxFileSizeMB()
			maxSizeBytes := maxSizeMB * 1024 * 1024
			oversizedContent := make([]byte, maxSizeBytes+1)

			t.Run("it should return 413 Payload Too Large", func(t *testing.T) {
				req := createMultipartRequest(oversizedContent, "test.txt")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusRequestEntityTooLarge {
					t.Errorf("expected status %d (413 Payload Too Large), got %d", http.StatusRequestEntityTooLarge, w.Code)
				}
			})
		})

		t.Run("and file size is within limit", func(t *testing.T) {
			setupRouter()
			validContent := []byte("small file content")

			t.Run("it should not reject based on size", func(t *testing.T) {
				mockEncryption.EXPECT().Encrypt(gomock.Any()).Return("encrypted", nil).AnyTimes()
				mockDataStore.EXPECT().WriteSecret(gomock.Any()).Return(nil).AnyTimes()

				req := createMultipartRequest(validContent, "test.txt")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code == http.StatusRequestEntityTooLarge {
					t.Errorf("expected status not to be %d (413 Payload Too Large), got %d", http.StatusRequestEntityTooLarge, w.Code)
				}
			})
		})
	})
}

func TestAccessSecretContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("when accessing a file secret", func(t *testing.T) {
		router := gin.New()
		cfg := settings.NewConfiguration()
		ctrl := gomock.NewController(t)
		mockDataStore := mocks.NewMockDataStore(ctrl)
		mockEncryption := mocks.NewMockEncryption(ctrl)

		router.Use(func(c *gin.Context) {
			c.Set(settings.Key, cfg)
			c.Set(datastore.Key, mockDataStore)
			c.Set(cryptography.Key, mockEncryption)
			c.Next()
		})

		router.POST("/v2/secrets/:id/access", AccessSecretContent)

		secret := &models.Secret{
			ID:          "test-id-123",
			Content:     []byte("file content"),
			ContentType: models.ContentTypeFile,
		}

		mockDataStore.EXPECT().ReadSecret("test-id-123").Return(secret)
		mockDataStore.EXPECT().IncreaseAccessCount("test-id-123").Return(int64(1), nil)
		mockEncryption.EXPECT().Decrypt(gomock.Any()).Return(secret.Content, nil)

		req, _ := http.NewRequest("POST", "/v2/secrets/test-id-123/access", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Run("it should include X-Content-Type-Options nosniff header", func(t *testing.T) {
			Equals(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		})

		t.Run("it should include Content-Security-Policy default-src none header", func(t *testing.T) {
			Equals(t, "default-src 'none'", w.Header().Get("Content-Security-Policy"))
		})

		t.Run("it should include X-Frame-Options DENY header", func(t *testing.T) {
			Equals(t, "DENY", w.Header().Get("X-Frame-Options"))
		})

		t.Run("it should include Cache-Control no-store header", func(t *testing.T) {
			Equals(t, "no-store, no-cache, must-revalidate", w.Header().Get("Cache-Control"))
		})
	})
}
