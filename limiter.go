package main

import (
	"sync"
	"time"
)

type clientState struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientState
	limit   int
	window  time.Duration
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
	Limit     int
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*clientState),
		limit:   limit,
		window:  window,
	}
}

func (rl *RateLimiter) Allow(apiKey string) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	state, exists := rl.clients[apiKey]
	if !exists {
		rl.clients[apiKey] = &clientState{
			count:       1,
			windowStart: now,
		}

		return RateLimitResult{
			Allowed:   true,
			Remaining: rl.limit - 1,
			ResetAt:   now.Add(rl.window),
			Limit:     rl.limit,
		}
	}

	if now.Sub(state.windowStart) >= rl.window {
		state.count = 1
		state.windowStart = now

		return RateLimitResult{
			Allowed:   true,
			Remaining: rl.limit - 1,
			ResetAt:   now.Add(rl.window),
			Limit:     rl.limit,
		}
	}

	if state.count >= rl.limit {
		return RateLimitResult{
			Allowed:   false,
			Remaining: 0,
			ResetAt:   state.windowStart.Add(rl.window),
			Limit:     rl.limit,
		}
	}

	state.count++
	return RateLimitResult{
		Allowed:   true,
		Remaining: rl.limit - state.count,
		ResetAt:   state.windowStart.Add(rl.window),
		Limit:     rl.limit,
	}
}
