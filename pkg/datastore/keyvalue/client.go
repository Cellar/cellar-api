package keyvalue

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../../mocks/mock_keyvalue_client.go -package=mocks . Client
type Client interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Incr(ctx context.Context, key string) (int64, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Info(ctx context.Context, section string) (string, error)
	Close() error
}

type Server struct {
	Name       string
	VersionKey string
}
