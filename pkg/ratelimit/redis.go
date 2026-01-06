package ratelimit

import (
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/settings"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type RedisRateLimiter struct {
	client        *redis.Client
	config        settings.IRateLimitConfiguration
	windowSeconds int
	logger        *log.Entry
}

func NewRedisRateLimiter(client *redis.Client, config settings.IRateLimitConfiguration) *RedisRateLimiter {
	logger := log.WithFields(log.Fields{
		"context": "ratelimiter",
		"backend": "redis",
	})

	logger.Debug("initializing Redis rate limiter")

	return &RedisRateLimiter{
		client:        client,
		config:        config,
		windowSeconds: config.WindowSeconds(),
		logger:        logger,
	}
}

func (rl *RedisRateLimiter) Allow(ctx context.Context, identifier string, tier Tier) (*Result, error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil, err
	}

	limit := rl.getLimitForTier(tier)
	key := rl.getKey(identifier, tier)
	now := time.Now()
	windowStart := now.Add(-time.Duration(rl.windowSeconds) * time.Second)

	pipe := rl.client.Pipeline()

	zremCmd := pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))
	zcardCmd := pipe.ZCard(ctx, key)
	zaddCmd := pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})
	expireCmd := pipe.Expire(ctx, key, time.Duration(rl.windowSeconds)*time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		rl.logger.WithError(err).WithField("identifier", identifier).Error("failed to execute rate limit pipeline")
		return nil, err
	}

	if err := zremCmd.Err(); err != nil {
		rl.logger.WithError(err).WithField("identifier", identifier).Error("failed to remove old entries")
		return nil, err
	}

	currentCount, err := zcardCmd.Result()
	if err != nil {
		rl.logger.WithError(err).WithField("identifier", identifier).Error("failed to count entries")
		return nil, err
	}

	if err := zaddCmd.Err(); err != nil {
		rl.logger.WithError(err).WithField("identifier", identifier).Error("failed to add new entry")
		return nil, err
	}

	if err := expireCmd.Err(); err != nil {
		rl.logger.WithError(err).WithField("identifier", identifier).Warn("failed to set expiration")
	}

	allowed := currentCount < int64(limit)
	remaining := limit - int(currentCount) - 1
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(time.Duration(rl.windowSeconds) * time.Second)
	retryAfter := 0
	if !allowed {
		retryAfter = rl.windowSeconds
	}

	result := &Result{
		Allowed:    allowed,
		Limit:      limit,
		Remaining:  remaining,
		RetryAfter: retryAfter,
		ResetAt:    resetAt,
	}

	if !allowed {
		rl.logger.WithFields(log.Fields{
			"identifier": identifier,
			"tier":       rl.getTierName(tier),
			"limit":      limit,
			"count":      currentCount,
		}).Warn("rate limit exceeded")
	}

	return result, nil
}

func (rl *RedisRateLimiter) getLimitForTier(tier Tier) int {
	switch tier {
	case Tier1:
		return rl.config.Tier1RequestsPerWindow()
	case Tier2:
		return rl.config.Tier2RequestsPerWindow()
	case Tier3:
		return rl.config.Tier3RequestsPerWindow()
	case HealthCheck:
		return rl.config.HealthCheckRequestsPerWindow()
	default:
		return rl.config.Tier3RequestsPerWindow()
	}
}

func (rl *RedisRateLimiter) getKey(identifier string, tier Tier) string {
	return fmt.Sprintf("cellar:ratelimit:%s:%s", identifier, rl.getTierName(tier))
}

func (rl *RedisRateLimiter) getTierName(tier Tier) string {
	switch tier {
	case Tier1:
		return "tier1"
	case Tier2:
		return "tier2"
	case Tier3:
		return "tier3"
	case HealthCheck:
		return "health"
	default:
		return "unknown"
	}
}
