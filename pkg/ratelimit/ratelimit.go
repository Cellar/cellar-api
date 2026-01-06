package ratelimit

import (
	"context"
	"time"
)

var Key = "RATE_LIMITER"

type Tier int

const (
	Tier1 Tier = iota
	Tier2
	Tier3
	HealthCheck
)

type Result struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter int
	ResetAt    time.Time
}

//go:generate mockgen -destination=../mocks/mock_ratelimiter.go -package=mocks . RateLimiter
type RateLimiter interface {
	Allow(ctx context.Context, identifier string, tier Tier) (*Result, error)
}
