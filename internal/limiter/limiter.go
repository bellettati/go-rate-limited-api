package limiter

import "time"

type Limiter interface {
	Allow(apiKey string) RateLimitResult
}

type LimitConfig struct {
	Limit  int
	Window time.Duration
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
	Limit     int
}
