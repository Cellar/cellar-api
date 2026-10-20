package ratelimit

import (
	"cellar/pkg/settings"
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisWindowStore struct {
	client *redis.Client
}

func NewRedisWindowStore(client *redis.Client) *RedisWindowStore {
	return &RedisWindowStore{client: client}
}

func NewRedisRateLimiter(client *redis.Client, config settings.IRateLimitConfiguration) *SlidingWindowRateLimiter {
	return NewSlidingWindowRateLimiter(NewRedisWindowStore(client), config)
}

func (s *RedisWindowStore) Hit(ctx context.Context, key string, now time.Time, window time.Duration) (int64, error) {
	pipe := s.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now.Add(-window).UnixNano(), 10))
	zcard := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixNano()), Member: now.UnixNano()})
	pipe.Expire(ctx, key, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return zcard.Val(), nil
}
