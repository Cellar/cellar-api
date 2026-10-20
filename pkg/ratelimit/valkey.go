package ratelimit

import (
	"cellar/pkg/settings"
	"context"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
)

type ValkeyWindowStore struct {
	client valkey.Client
}

func NewValkeyWindowStore(client valkey.Client) *ValkeyWindowStore {
	return &ValkeyWindowStore{client: client}
}

func NewValkeyRateLimiter(client valkey.Client, config settings.IRateLimitConfiguration) *SlidingWindowRateLimiter {
	return NewSlidingWindowRateLimiter(NewValkeyWindowStore(client), config)
}

func (s *ValkeyWindowStore) Hit(ctx context.Context, key string, now time.Time, window time.Duration) (int64, error) {
	nowNano := now.UnixNano()
	b := s.client.B()
	results := s.client.DoMulti(ctx,
		b.Zremrangebyscore().Key(key).Min("0").Max(strconv.FormatInt(now.Add(-window).UnixNano(), 10)).Build(),
		b.Zcard().Key(key).Build(),
		b.Zadd().Key(key).ScoreMember().ScoreMember(float64(nowNano), strconv.FormatInt(nowNano, 10)).Build(),
		b.Expire().Key(key).Seconds(int64(window/time.Second)).Build(),
	)

	for _, result := range results {
		if err := result.Error(); err != nil {
			return 0, err
		}
	}
	return results[1].AsInt64()
}
