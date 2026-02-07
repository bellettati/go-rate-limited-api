package main

import "time"

type Limiter interface {
	Allow(apiKey string) RateLimitResult
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
	Limit     int
}
