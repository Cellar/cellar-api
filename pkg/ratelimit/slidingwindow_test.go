package ratelimit_test

import (
	"cellar/pkg/mocks"
	"cellar/pkg/ratelimit"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const window = 60

func newLimiter(t *testing.T) (*ratelimit.SlidingWindowRateLimiter, *mocks.MockWindowStore) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockWindowStore(ctrl)
	config := mocks.NewMockIRateLimitConfiguration(ctrl)
	config.EXPECT().WindowSeconds().Return(window).AnyTimes()
	config.EXPECT().Tier1RequestsPerWindow().Return(10).AnyTimes()
	config.EXPECT().Tier2RequestsPerWindow().Return(20).AnyTimes()
	config.EXPECT().Tier3RequestsPerWindow().Return(30).AnyTimes()
	config.EXPECT().HealthCheckRequestsPerWindow().Return(40).AnyTimes()
	return ratelimit.NewSlidingWindowRateLimiter(store, config), store
}

func TestSlidingWindowRateLimiter(t *testing.T) {
	ctx := context.Background()

	t.Run("when checking a tier", func(t *testing.T) {
		tiers := []struct {
			tier  ratelimit.Tier
			key   string
			limit int
		}{
			{ratelimit.Tier1, "cellar:ratelimit:client:tier1", 10},
			{ratelimit.Tier2, "cellar:ratelimit:client:tier2", 20},
			{ratelimit.Tier3, "cellar:ratelimit:client:tier3", 30},
			{ratelimit.HealthCheck, "cellar:ratelimit:client:health", 40},
		}

		for _, tc := range tiers {
			t.Run("and the tier uses key "+tc.key, func(t *testing.T) {
				limiter, store := newLimiter(t)
				store.EXPECT().Hit(ctx, tc.key, gomock.Any(), window*time.Second).Return(int64(0), nil)
				result, _ := limiter.Allow(ctx, "client", tc.tier)

				t.Run("it should apply the tier limit", func(t *testing.T) {
					assert.Equal(t, tc.limit, result.Limit)
				})
			})
		}
	})

	t.Run("when the window has room", func(t *testing.T) {
		limiter, store := newLimiter(t)
		store.EXPECT().Hit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(3), nil)
		result, _ := limiter.Allow(ctx, "client", ratelimit.Tier1)

		expectations := []struct {
			name     string
			actual   any
			expected any
		}{
			{"allow the request", result.Allowed, true},
			{"report the remaining capacity", result.Remaining, 6},
			{"have zero retry after", result.RetryAfter, 0},
		}
		for _, e := range expectations {
			t.Run("it should "+e.name, func(t *testing.T) {
				assert.Equal(t, e.expected, e.actual)
			})
		}
	})

	t.Run("when the window is full", func(t *testing.T) {
		limiter, store := newLimiter(t)
		store.EXPECT().Hit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(10), nil)
		result, _ := limiter.Allow(ctx, "client", ratelimit.Tier1)

		expectations := []struct {
			name     string
			actual   any
			expected any
		}{
			{"deny the request", result.Allowed, false},
			{"report zero remaining", result.Remaining, 0},
			{"ask the client to retry after the window", result.RetryAfter, window},
		}
		for _, e := range expectations {
			t.Run("it should "+e.name, func(t *testing.T) {
				assert.Equal(t, e.expected, e.actual)
			})
		}
	})

	t.Run("when the store fails", func(t *testing.T) {
		limiter, store := newLimiter(t)
		boom := errors.New("boom")
		store.EXPECT().Hit(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), boom)
		_, err := limiter.Allow(ctx, "client", ratelimit.Tier1)

		t.Run("it should return the error", func(t *testing.T) {
			assert.ErrorIs(t, err, boom)
		})
	})

	t.Run("when the context is cancelled", func(t *testing.T) {
		limiter, _ := newLimiter(t)
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		_, err := limiter.Allow(cancelled, "client", ratelimit.Tier1)

		t.Run("it should return an error", func(t *testing.T) {
			assert.Error(t, err)
		})
	})
}
