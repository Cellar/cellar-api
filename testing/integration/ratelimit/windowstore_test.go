//go:build integration

package ratelimit

import (
	"cellar/pkg/mocks"
	"cellar/pkg/ratelimit"
	"cellar/testing/testhelpers"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newLimiter(t *testing.T, backend testhelpers.Backend, windowSeconds, tier1, tier2 int) *ratelimit.SlidingWindowRateLimiter {
	ctrl := gomock.NewController(t)
	config := mocks.NewMockIRateLimitConfiguration(ctrl)
	config.EXPECT().WindowSeconds().Return(windowSeconds).AnyTimes()
	config.EXPECT().Tier1RequestsPerWindow().Return(tier1).AnyTimes()
	config.EXPECT().Tier2RequestsPerWindow().Return(tier2).AnyTimes()
	return ratelimit.NewSlidingWindowRateLimiter(backend.WindowStore, config)
}

func exhaust(t *testing.T, limiter ratelimit.RateLimiter, identifier string, tier ratelimit.Tier, times int) {
	for i := 0; i < times; i++ {
		_, err := limiter.Allow(context.Background(), identifier, tier)
		require.NoError(t, err)
	}
}

func TestWindowStore(t *testing.T) {
	testhelpers.ForEachBackend(t, func(t *testing.T, backend testhelpers.Backend) {
		ctx := context.Background()

		t.Run("when a client is within the limit", func(t *testing.T) {
			limiter := newLimiter(t, backend, 60, 3, 3)
			identifier := testhelpers.RandomId(t)
			result, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
			require.NoError(t, err)

			t.Run("it should allow the request", func(t *testing.T) {
				assert.True(t, result.Allowed)
			})
		})

		t.Run("when a client reaches the limit", func(t *testing.T) {
			limiter := newLimiter(t, backend, 60, 3, 3)
			identifier := testhelpers.RandomId(t)
			exhaust(t, limiter, identifier, ratelimit.Tier1, 3)
			result, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
			require.NoError(t, err)

			t.Run("it should deny the request", func(t *testing.T) {
				assert.False(t, result.Allowed)
			})

			t.Run("and a different client makes a request", func(t *testing.T) {
				other, err := limiter.Allow(ctx, testhelpers.RandomId(t), ratelimit.Tier1)
				require.NoError(t, err)

				t.Run("it should allow the other client", func(t *testing.T) {
					assert.True(t, other.Allowed)
				})
			})

			t.Run("and the same client uses a different tier", func(t *testing.T) {
				other, err := limiter.Allow(ctx, identifier, ratelimit.Tier2)
				require.NoError(t, err)

				t.Run("it should allow the other tier", func(t *testing.T) {
					assert.True(t, other.Allowed)
				})
			})
		})

		t.Run("when the window passes", func(t *testing.T) {
			limiter := newLimiter(t, backend, 1, 2, 2)
			identifier := testhelpers.RandomId(t)
			exhaust(t, limiter, identifier, ratelimit.Tier1, 3)
			time.Sleep(1100 * time.Millisecond)
			result, err := limiter.Allow(ctx, identifier, ratelimit.Tier1)
			require.NoError(t, err)

			t.Run("it should allow new requests", func(t *testing.T) {
				assert.True(t, result.Allowed)
			})
		})

		t.Run("when recording a hit", func(t *testing.T) {
			key := "cellar:ratelimit:" + testhelpers.RandomId(t) + ":tier1"
			t.Cleanup(func() { _ = backend.Raw.Del(ctx, key).Err() })
			_, err := backend.WindowStore.Hit(ctx, key, time.Now(), 30*time.Second)
			require.NoError(t, err)

			t.Run("it should expire the window key with the window", func(t *testing.T) {
				ttl, err := backend.Raw.TTL(ctx, key).Result()
				require.NoError(t, err)
				assert.InDelta(t, 30*time.Second, ttl, float64(2*time.Second))
			})
		})
	})
}
