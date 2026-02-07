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
	mu           sync.Mutex
	clients      map[string]*clientState
	defaultLimit LimitConfig
	overrides    map[string]LimitConfig
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
	Limit     int
}

type LimitConfig struct {
	Limit  int
	Window time.Duration
}

func NewRateLimiter(defaultLimit LimitConfig, overrides map[string]LimitConfig) *RateLimiter {
	if overrides == nil {
		overrides = make(map[string]LimitConfig)
	}

	return &RateLimiter{
		clients:      make(map[string]*clientState),
		defaultLimit: defaultLimit,
		overrides:    overrides,
	}
}

func (rl *RateLimiter) Allow(apiKey string) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	state, exists := rl.clients[apiKey]

	cfg, ok := rl.overrides[apiKey]
	if !ok {
		cfg = rl.defaultLimit
	}

	if !exists {
		rl.clients[apiKey] = &clientState{
			count:       1,
			windowStart: now,
		}

		return RateLimitResult{
			Allowed:   true,
			Remaining: cfg.Limit - 1,
			ResetAt:   now.Add(cfg.Window),
			Limit:     cfg.Limit,
		}
	}

	if now.Sub(state.windowStart) >= cfg.Window {
		state.count = 1
		state.windowStart = now

		return RateLimitResult{
			Allowed:   true,
			Remaining: cfg.Limit - 1,
			ResetAt:   now.Add(cfg.Window),
			Limit:     cfg.Limit,
		}
	}

	if state.count >= cfg.Limit {
		return RateLimitResult{
			Allowed:   false,
			Remaining: 0,
			ResetAt:   state.windowStart.Add(cfg.Window),
			Limit:     cfg.Limit,
		}
	}

	state.count++

	return RateLimitResult{
		Allowed:   true,
		Remaining: cfg.Limit - state.count,
		ResetAt:   state.windowStart.Add(cfg.Window),
		Limit:     cfg.Limit,
	}
}
