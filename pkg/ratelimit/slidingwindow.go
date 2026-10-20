package ratelimit

import (
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/settings"
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type SlidingWindowRateLimiter struct {
	store         WindowStore
	config        settings.IRateLimitConfiguration
	windowSeconds int
	logger        *log.Entry
}

func NewSlidingWindowRateLimiter(store WindowStore, config settings.IRateLimitConfiguration) *SlidingWindowRateLimiter {
	logger := log.WithFields(log.Fields{
		"context": "ratelimiter",
	})

	logger.Debug("initializing sliding window rate limiter")

	return &SlidingWindowRateLimiter{
		store:         store,
		config:        config,
		windowSeconds: config.WindowSeconds(),
		logger:        logger,
	}
}

func (rl *SlidingWindowRateLimiter) Allow(ctx context.Context, identifier string, tier Tier) (*Result, error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return nil, err
	}

	limit := rl.getLimitForTier(tier)
	key := rl.getKey(identifier, tier)
	window := time.Duration(rl.windowSeconds) * time.Second
	now := time.Now()

	currentCount, err := rl.store.Hit(ctx, key, now, window)
	if err != nil {
		rl.logger.WithError(err).WithField("identifier", identifier).Error("failed to record rate limit hit")
		return nil, err
	}

	allowed := currentCount < int64(limit)
	remaining := limit - int(currentCount) - 1
	if remaining < 0 {
		remaining = 0
	}

	retryAfter := 0
	if !allowed {
		retryAfter = rl.windowSeconds
		rl.logger.WithFields(log.Fields{
			"identifier": identifier,
			"tier":       rl.getTierName(tier),
			"limit":      limit,
			"count":      currentCount,
		}).Warn("rate limit exceeded")
	}

	return &Result{
		Allowed:    allowed,
		Limit:      limit,
		Remaining:  remaining,
		RetryAfter: retryAfter,
		ResetAt:    now.Add(window),
	}, nil
}

func (rl *SlidingWindowRateLimiter) getLimitForTier(tier Tier) int {
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

func (rl *SlidingWindowRateLimiter) getKey(identifier string, tier Tier) string {
	return fmt.Sprintf("cellar:ratelimit:%s:%s", identifier, rl.getTierName(tier))
}

func (rl *SlidingWindowRateLimiter) getTierName(tier Tier) string {
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
