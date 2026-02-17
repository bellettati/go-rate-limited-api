package limiter

import (
	"sync"
	"time"
)

const slidingWindowCapMax = 256

type slidingWindowState struct {
	timestamps  []time.Time
	lastSeen time.Time
}

type SlidingWindowLimiter struct {
	mu           sync.Mutex
	clients      map[string]*slidingWindowState
	defaultLimit LimitConfig
	overrides    map[string]LimitConfig
	clock Clock
}

func NewSlidingWindowLimiter(clock Clock, defaultLimit LimitConfig, overrides map[string]LimitConfig) *SlidingWindowLimiter {
	if overrides == nil {
		overrides = make(map[string]LimitConfig)
	}

	sw := &SlidingWindowLimiter{
		clients:      make(map[string]*slidingWindowState),
		defaultLimit: defaultLimit,
		overrides:    overrides,
		clock: clock,
	}

	go sw.startCleanup()

	return sw 
}

func (sw *SlidingWindowLimiter) configFor(apiKey string) LimitConfig {
	if cfg, ok := sw.overrides[apiKey]; ok {
		return cfg
	}

	return sw.defaultLimit
}

func (sw *SlidingWindowLimiter) cleanup() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := sw.clock.Now()

	for key, client := range sw.clients {
		cfg := sw.configFor(key)
		if now.Sub(client.lastSeen) > cfg.Window {
			delete(sw.clients, key)
		}
	}
}

func (sw *SlidingWindowLimiter) startCleanup() {
	ticker := time.NewTicker(time.Minute)

	for range ticker.C {
		sw.cleanup()
	}
}

func (sw *SlidingWindowLimiter) Allow(apiKey string) RateLimitResult {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := sw.clock.Now()
	cfg := sw.configFor(apiKey)

	state, exists := sw.clients[apiKey]
	if !exists {
		capHint := cfg.Limit
		if capHint > slidingWindowCapMax {
			capHint = slidingWindowCapMax
		}
		if capHint < 0 {
			capHint = 0
		}
		state = &slidingWindowState{
			timestamps: make([]time.Time, 0, capHint),
			lastSeen: now,
		}
		sw.clients[apiKey] = state
	}
	state.lastSeen = now

	windowStart := now.Add(-cfg.Window)

	valid := state.timestamps[:0]
	for _, ts := range state.timestamps {
		if !ts.Before(windowStart) {
			valid = append(valid, ts)
		}
	}
	state.timestamps = valid

	if len(state.timestamps) >= cfg.Limit {
		return RateLimitResult{
			Allowed: false,
			Remaining: 0,
			ResetAt: state.timestamps[0].Add(cfg.Window),
			Limit: cfg.Limit,
		}
	}

	state.timestamps = append(state.timestamps, now)
	return RateLimitResult{
		Allowed: true,
		Remaining: cfg.Limit - len(state.timestamps),
		ResetAt: state.timestamps[0].Add(cfg.Window),
		Limit: cfg.Limit,
	}
}
