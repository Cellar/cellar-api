//go:build integration

package middleware

import (
	"cellar/pkg/middleware"
	"cellar/pkg/mocks"
	"cellar/pkg/ratelimit"
	"cellar/pkg/settings"
	"cellar/testing/testhelpers"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var ipCounter uint32

func getUniqueIP() string {
	counter := atomic.AddUint32(&ipCounter, 1)
	return fmt.Sprintf("192.168.%d.%d:12345", (counter/256)%256, counter%256)
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("when rate limit is disabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client, config := setupRateLimitTest(t)
		config.EXPECT().WindowSeconds().Return(60).AnyTimes()
		config.EXPECT().Enabled().Return(false).AnyTimes()

		router := setupRouter(ctrl, client, config)

		t.Run("it should not apply rate limiting", func(t *testing.T) {
			clientIP := getUniqueIP()
			for i := 0; i < 20; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = clientIP
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	})

	t.Run("when rate limit is enabled", func(t *testing.T) {
		t.Run("and within limit", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client, config := setupRateLimitTest(t)
			config.EXPECT().Enabled().Return(true).AnyTimes()
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(5).AnyTimes()

			router := setupRouter(ctrl, client, config)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = getUniqueIP()
			router.ServeHTTP(w, req)

			t.Run("it should allow the request", func(t *testing.T) {
				assert.Equal(t, http.StatusOK, w.Code)
			})

			t.Run("it should include rate limit headers", func(t *testing.T) {
				assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
			})

			t.Run("it should include remaining count", func(t *testing.T) {
				remaining := w.Header().Get("X-RateLimit-Remaining")
				assert.NotEmpty(t, remaining)
			})

			t.Run("it should include reset timestamp", func(t *testing.T) {
				assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
			})
		})

		t.Run("and limit exceeded", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client, config := setupRateLimitTest(t)
			config.EXPECT().Enabled().Return(true).AnyTimes()
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(3).AnyTimes()

			router := setupRouter(ctrl, client, config)

			clientIP := getUniqueIP()
			for i := 0; i < 3; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = clientIP
				router.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = clientIP
			router.ServeHTTP(w, req)

			t.Run("it should return 429 status", func(t *testing.T) {
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			})

			t.Run("it should have zero remaining", func(t *testing.T) {
				remaining := w.Header().Get("X-RateLimit-Remaining")
				remainingInt, err := strconv.Atoi(remaining)
				require.NoError(t, err)
				assert.Equal(t, 0, remainingInt)
			})

			t.Run("it should include retry-after header", func(t *testing.T) {
				assert.NotEmpty(t, w.Header().Get("Retry-After"))
			})
		})

		t.Run("and different client IPs", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client, config := setupRateLimitTest(t)
			config.EXPECT().Enabled().Return(true).AnyTimes()
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(2).AnyTimes()

			router := setupRouter(ctrl, client, config)

			clientA := getUniqueIP()
			for i := 0; i < 2; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = clientA
				router.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)
			}

			t.Run("it should block client A", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = clientA
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			})

			t.Run("it should allow client B independently", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = getUniqueIP()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			})
		})

		t.Run("and different tiers", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client, config := setupRateLimitTest(t)
			config.EXPECT().Enabled().Return(true).AnyTimes()
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(2).AnyTimes()
			config.EXPECT().Tier2RequestsPerWindow().Return(5).AnyTimes()

			router := setupRouterWithTiers(ctrl, client, config)

			clientIP := getUniqueIP()
			for i := 0; i < 2; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/tier1", nil)
				req.RemoteAddr = clientIP
				router.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)
			}

			t.Run("it should block tier1 when limit reached", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/tier1", nil)
				req.RemoteAddr = clientIP
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			})

			t.Run("it should allow tier2 independently", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/tier2", nil)
				req.RemoteAddr = clientIP
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			})
		})

		t.Run("and using X-Forwarded-For header", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client, config := setupRateLimitTest(t)
			config.EXPECT().Enabled().Return(true).AnyTimes()
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(2).AnyTimes()

			router := setupRouter(ctrl, client, config)

			forwardedIP := getUniqueIP()
			for i := 0; i < 2; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Forwarded-For", forwardedIP)
				router.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)
			}

			t.Run("it should use forwarded IP for rate limiting", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Forwarded-For", forwardedIP)
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			})
		})
	})
}

func setupRateLimitTest(t *testing.T) (*redis.Client, *mocks.MockIRateLimitConfiguration) {
	cfg := settings.NewConfiguration()
	client := testhelpers.GetRedisClient(cfg.Datastore().Redis())

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Redis must be available for integration tests: %v", err)
	}

	t.Cleanup(func() {
		client.FlushDB(ctx)
		_ = client.Close()
	})

	ctrl := gomock.NewController(t)
	config := mocks.NewMockIRateLimitConfiguration(ctrl)

	return client, config
}

func setupRouter(ctrl *gomock.Controller, client *redis.Client, config settings.IRateLimitConfiguration) *gin.Engine {
	router := gin.New()
	router.Use(middleware.ErrorHandler())

	limiter := ratelimit.NewRedisRateLimiter(client, config)

	mockConfig := mocks.NewMockIConfiguration(ctrl)
	mockConfig.EXPECT().RateLimit().Return(config).AnyTimes()

	router.Use(func(c *gin.Context) {
		c.Set(settings.Key, mockConfig)
		c.Set(ratelimit.Key, limiter)
		c.Next()
	})

	router.GET("/test", middleware.RateLimit(ratelimit.Tier1), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	return router
}

func setupRouterWithTiers(ctrl *gomock.Controller, client *redis.Client, config settings.IRateLimitConfiguration) *gin.Engine {
	router := gin.New()
	router.Use(middleware.ErrorHandler())

	limiter := ratelimit.NewRedisRateLimiter(client, config)

	mockConfig := mocks.NewMockIConfiguration(ctrl)
	mockConfig.EXPECT().RateLimit().Return(config).AnyTimes()

	router.Use(func(c *gin.Context) {
		c.Set(settings.Key, mockConfig)
		c.Set(ratelimit.Key, limiter)
		c.Next()
	})

	router.GET("/tier1", middleware.RateLimit(ratelimit.Tier1), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "tier1"})
	})

	router.GET("/tier2", middleware.RateLimit(ratelimit.Tier2), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "tier2"})
	})

	return router
}
