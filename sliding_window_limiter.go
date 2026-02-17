package main

import (
	"sync"
	"time"
)

type slidingWindowState struct {
	timestamps  []time.Time
}

type SlidingWindowLimiter struct {
	mu           sync.Mutex
	clients      map[string]*slidingWindowState
	defaultLimit LimitConfig
	overrides    map[string]LimitConfig
}

func NewSlidingWindowLimiter(defaultLimit LimitConfig, overrides map[string]LimitConfig) *SlidingWindowLimiter {
	if overrides == nil {
		overrides = make(map[string]LimitConfig)
	}

	return &SlidingWindowLimiter{
		clients:      make(map[string]*slidingWindowState),
		defaultLimit: defaultLimit,
		overrides:    overrides,
	}
}

func (sw *SlidingWindowLimiter) configFor(apiKey string) LimitConfig {
	if cfg, ok := sw.overrides[apiKey]; ok {
		return cfg
	}

	return sw.defaultLimit
}

func (sw *SlidingWindowLimiter) Allow(apiKey string) RateLimitResult {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cfg := sw.configFor(apiKey)

	state, exists := sw.clients[apiKey]
	if !exists {
		state = &slidingWindowState{
			timestamps: make([]time.Time, 0, cfg.Limit),
		}
		sw.clients[apiKey] = state
	}

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
