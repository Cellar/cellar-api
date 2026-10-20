package ratelimit

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_windowstore.go -package=mocks . WindowStore
type WindowStore interface {
	Hit(ctx context.Context, key string, now time.Time, window time.Duration) (int64, error)
}
