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

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*clientState),
		limit:   limit,
		window:  window,
	}
}

func (rl *RateLimiter) Allow(apiKey string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	state, exists := rl.clients[apiKey]
	if !exists {
		rl.clients[apiKey] = &clientState{
			count:       1,
			windowStart: now,
		}

		return true
	}

	if now.Sub(state.windowStart) >= rl.window {
		state.count = 1
		state.windowStart = now
		return true
	}

	if state.count >= rl.limit {
		return false
	}

	state.count++
	return true
}
