package store

import (
	"context"
	"time"
)

type Store interface {
	IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (value int64, ttlRemaining time.Duration, err error)

	Close() error
}