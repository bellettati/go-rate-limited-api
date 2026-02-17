package limiter

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
	clock Clock
}

func NewTokenBucketLimiter(
	clock Clock,
	defaultLimit LimitConfig,
	overrides map[string]LimitConfig,
) *TokenBucketLimiter {
	if overrides == nil {
		overrides = make(map[string]LimitConfig)
	}

	tb := &TokenBucketLimiter{
		clients:      make(map[string]*tokenBucketState),
		defaultLimit: defaultLimit,
		overrides:    overrides,
		clock: clock,
	}

	go tb.startCleanup()

	return tb
}

func (tb *TokenBucketLimiter) configFor(apiKey string) LimitConfig {
	if cfg, ok := tb.overrides[apiKey]; ok {
		return cfg
	}
	return tb.defaultLimit
}

func (tb *TokenBucketLimiter) cleanup() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := tb.clock.Now()

	for key, client := range tb.clients {
		cfg := tb.configFor(key)
		if now.Sub(client.lastRefill) > cfg.Window {
			delete(tb.clients, key)
		} 
	}
}

func (tb *TokenBucketLimiter) startCleanup() {
	ticker := time.NewTicker(time.Minute)

	for range ticker.C {
		tb.cleanup()
	}
}

func (tb *TokenBucketLimiter) Allow(apiKey string) RateLimitResult {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	cfg := tb.configFor(apiKey)
	now := tb.clock.Now()

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
