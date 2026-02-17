package limiter

import (
	"sync"
	"time"
)

type clientState struct {
	count       int
	windowStart time.Time
	lastSeen time.Time
}

type FixedWindowLimiter struct {
	mu           sync.Mutex
	clients      map[string]*clientState
	defaultLimit LimitConfig
	overrides    map[string]LimitConfig
	clock Clock
}

func NewFixedWindowLimiter(clock Clock, defaultLimit LimitConfig, overrides map[string]LimitConfig) *FixedWindowLimiter {
	if overrides == nil {
		overrides = make(map[string]LimitConfig)
	}

	rl := &FixedWindowLimiter{
		clients:      make(map[string]*clientState),
		defaultLimit: defaultLimit,
		overrides:    overrides,
		clock: clock,
	}

	go rl.startCleanup()

	return rl 
}

func (rl *FixedWindowLimiter) configFor(apiKey string) LimitConfig {
	if cfg, ok := rl.overrides[apiKey]; ok {
		return cfg
	}

	return rl.defaultLimit
}

func (rl *FixedWindowLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.clock.Now()

	for key, client := range rl.clients {
		cfg := rl.configFor(key)
		if now.Sub(client.lastSeen) > cfg.Window {
			delete(rl.clients, key)
		}
	}
}

func (rl *FixedWindowLimiter) startCleanup() {
	ticker := time.NewTicker(time.Minute)

	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *FixedWindowLimiter) Allow(apiKey string) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.clock.Now()

	state, exists := rl.clients[apiKey]

	cfg := rl.configFor(apiKey)

	if !exists {
		rl.clients[apiKey] = &clientState{
			count:       1,
			windowStart: now,
			lastSeen: now,
		}

		return RateLimitResult{
			Allowed:   true,
			Remaining: cfg.Limit - 1,
			ResetAt:   now.Add(cfg.Window),
			Limit:     cfg.Limit,
		}
	}

	state.lastSeen = now

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
