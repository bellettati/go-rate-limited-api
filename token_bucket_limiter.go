package main

import (
	"sync"
	"time"
)

type tokenBucketState struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucketLimiter struct {
	mu           sync.Mutex
	clients      map[string]*tokenBucketState
	defaultLimit LimitConfig
	overrides    map[string]LimitConfig
}

func NewTokenBucketLimiter(
	defaultLimit LimitConfig,
	overrides map[string]LimitConfig,
) *TokenBucketLimiter {
	if overrides == nil {
		overrides = make(map[string]LimitConfig)
	}

	return &TokenBucketLimiter{
		clients:      make(map[string]*tokenBucketState),
		defaultLimit: defaultLimit,
		overrides:    overrides,
	}
}

func (tb *TokenBucketLimiter) limitFor(apiKey string) LimitConfig {
	if cfg, ok := tb.overrides[apiKey]; ok {
		return cfg
	}
	return tb.defaultLimit
}

func (tb *TokenBucketLimiter) Allow(apiKey string) RateLimitResult {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	cfg := tb.limitFor(apiKey)
	now := time.Now()

	state, exists := tb.clients[apiKey]
	if !exists {
		tb.clients[apiKey] = &tokenBucketState{
			tokens:     float64(cfg.Limit - 1),
			lastRefill: now,
		}

		return RateLimitResult{
			Allowed:   true,
			Remaining: cfg.Limit - 1,
			Limit:     cfg.Limit,
			ResetAt:   now.Add(cfg.Window),
		}
	}

	elapsed := now.Sub(state.lastRefill)
	refillRate := float64(cfg.Limit) / cfg.Window.Seconds()
	refilled := refillRate * elapsed.Seconds()

	state.tokens = min(state.tokens+refilled, float64(cfg.Limit))
	state.lastRefill = now

	if state.tokens < 1 {
		return RateLimitResult{
			Allowed:   false,
			Remaining: 0,
			Limit:     cfg.Limit,
			ResetAt:   now.Add(cfg.Window),
		}
	}

	state.tokens--

	return RateLimitResult{
		Allowed:   true,
		Remaining: int(state.tokens),
		Limit:     cfg.Limit,
		ResetAt:   now.Add(cfg.Window),
	}
}
