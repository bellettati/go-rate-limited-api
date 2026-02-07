package main

import "time"

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
	Limit     int
}

type Limiter interface {
	Allow(apiKey string) RateLimitResult
}
