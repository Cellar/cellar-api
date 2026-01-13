package ratelimit_test

import (
	"cellar/pkg/mocks"
	"cellar/pkg/ratelimit"
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRedisRateLimiter(t *testing.T) {
	t.Run("when checking rate limit", func(t *testing.T) {
		t.Run("and within limit", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := setupRedisClient(t)
			config := mocks.NewMockIRateLimitConfiguration(ctrl)
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(10).AnyTimes()

			limiter := ratelimit.NewRedisRateLimiter(client, config)

			t.Run("it should allow the request", func(t *testing.T) {
				result, err := limiter.Allow(context.Background(), "test-client", ratelimit.Tier1)

				require.NoError(t, err)
				assert.True(t, result.Allowed)
			})

			t.Run("it should return correct limit", func(t *testing.T) {
				result, err := limiter.Allow(context.Background(), "test-client-2", ratelimit.Tier1)

				require.NoError(t, err)
				assert.Equal(t, 10, result.Limit)
			})

			t.Run("it should have remaining capacity", func(t *testing.T) {
				result, err := limiter.Allow(context.Background(), "test-client-3", ratelimit.Tier1)

				require.NoError(t, err)
				assert.GreaterOrEqual(t, result.Remaining, 0)
			})

			t.Run("it should have zero retry after", func(t *testing.T) {
				result, err := limiter.Allow(context.Background(), "test-client-4", ratelimit.Tier1)

				require.NoError(t, err)
				assert.Equal(t, 0, result.RetryAfter)
			})
		})

		t.Run("and limit exceeded", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := setupRedisClient(t)
			config := mocks.NewMockIRateLimitConfiguration(ctrl)
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(10).AnyTimes()

			limiter := ratelimit.NewRedisRateLimiter(client, config)
			identifier := "test-client-exceeded"
			ctx := context.Background()

			for i := 0; i < 10; i++ {
				_, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
				require.NoError(t, err)
			}

			result, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
			require.NoError(t, err)

			t.Run("it should deny the request", func(t *testing.T) {
				assert.False(t, result.Allowed)
			})

			t.Run("it should have zero remaining", func(t *testing.T) {
				assert.Equal(t, 0, result.Remaining)
			})

			t.Run("it should have positive retry after", func(t *testing.T) {
				assert.Greater(t, result.RetryAfter, 0)
			})
		})

		t.Run("and using different tiers", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := setupRedisClient(t)
			config := mocks.NewMockIRateLimitConfiguration(ctrl)
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(300).AnyTimes()
			config.EXPECT().Tier2RequestsPerWindow().Return(600).AnyTimes()
			config.EXPECT().Tier3RequestsPerWindow().Return(1200).AnyTimes()
			config.EXPECT().HealthCheckRequestsPerWindow().Return(1200).AnyTimes()

			limiter := ratelimit.NewRedisRateLimiter(client, config)
			ctx := context.Background()

			t.Run("it should apply tier1 limit", func(t *testing.T) {
				result, err := limiter.Allow(ctx, "client-tier1", ratelimit.Tier1)

				require.NoError(t, err)
				assert.Equal(t, 300, result.Limit)
			})

			t.Run("it should apply tier2 limit", func(t *testing.T) {
				result, err := limiter.Allow(ctx, "client-tier2", ratelimit.Tier2)

				require.NoError(t, err)
				assert.Equal(t, 600, result.Limit)
			})

			t.Run("it should apply tier3 limit", func(t *testing.T) {
				result, err := limiter.Allow(ctx, "client-tier3", ratelimit.Tier3)

				require.NoError(t, err)
				assert.Equal(t, 1200, result.Limit)
			})

			t.Run("it should apply health check limit", func(t *testing.T) {
				result, err := limiter.Allow(ctx, "client-health", ratelimit.HealthCheck)

				require.NoError(t, err)
				assert.Equal(t, 1200, result.Limit)
			})
		})

		t.Run("and context is cancelled", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := setupRedisClient(t)
			config := mocks.NewMockIRateLimitConfiguration(ctrl)
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			limiter := ratelimit.NewRedisRateLimiter(client, config)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			t.Run("it should return context error", func(t *testing.T) {
				_, err := limiter.Allow(ctx, "test-client", ratelimit.Tier1)

				assert.Error(t, err)
			})
		})

		t.Run("and requests from different clients", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := setupRedisClient(t)
			config := mocks.NewMockIRateLimitConfiguration(ctrl)
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(10).AnyTimes()

			limiter := ratelimit.NewRedisRateLimiter(client, config)
			ctx := context.Background()

			for i := 0; i < 10; i++ {
				_, err := limiter.Allow(ctx, "client-a", ratelimit.Tier1)
				require.NoError(t, err)
			}

			t.Run("it should deny client A when limit reached", func(t *testing.T) {
				result, err := limiter.Allow(ctx, "client-a", ratelimit.Tier1)

				require.NoError(t, err)
				assert.False(t, result.Allowed)
			})

			t.Run("it should allow client B independently", func(t *testing.T) {
				result, err := limiter.Allow(ctx, "client-b", ratelimit.Tier1)

				require.NoError(t, err)
				assert.True(t, result.Allowed)
			})
		})

		t.Run("and requests to different tiers from same client", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := setupRedisClient(t)
			config := mocks.NewMockIRateLimitConfiguration(ctrl)
			config.EXPECT().WindowSeconds().Return(60).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(10).AnyTimes()
			config.EXPECT().Tier2RequestsPerWindow().Return(30).AnyTimes()

			limiter := ratelimit.NewRedisRateLimiter(client, config)
			ctx := context.Background()
			identifier := "same-client"

			for i := 0; i < 10; i++ {
				_, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
				require.NoError(t, err)
			}

			t.Run("it should deny tier1 when limit reached", func(t *testing.T) {
				result, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)

				require.NoError(t, err)
				assert.False(t, result.Allowed)
			})

			t.Run("it should allow tier2 independently", func(t *testing.T) {
				result, err := limiter.Allow(ctx, identifier, ratelimit.Tier2)

				require.NoError(t, err)
				assert.True(t, result.Allowed)
			})
		})
	})

	t.Run("when window resets", func(t *testing.T) {
		t.Run("and old requests expire", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := setupRedisClient(t)
			config := mocks.NewMockIRateLimitConfiguration(ctrl)
			config.EXPECT().WindowSeconds().Return(1).AnyTimes()
			config.EXPECT().Tier1RequestsPerWindow().Return(2).AnyTimes()

			limiter := ratelimit.NewRedisRateLimiter(client, config)
			ctx := context.Background()
			identifier := "test-window-reset"

			_, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
			require.NoError(t, err)

			_, err = limiter.Allow(ctx, identifier, ratelimit.Tier1)
			require.NoError(t, err)

			result, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
			require.NoError(t, err)
			assert.False(t, result.Allowed)

			time.Sleep(1100 * time.Millisecond)

			t.Run("it should allow new requests after window", func(t *testing.T) {
				result, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)

				require.NoError(t, err)
				assert.True(t, result.Allowed)
			})
		})
	})
}

func setupRedisClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	t.Cleanup(func() {
		client.FlushDB(ctx)
		_ = client.Close()
	})

	return client
}
