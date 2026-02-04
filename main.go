package main

import (
	"fmt"
	"log"
	"net/http"
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

func main() {
	mux := http.NewServeMux()
	limiter := NewRateLimiter(10, time.Minute)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")

		if apiKey == "" {
			http.Error(w, "missing API key", http.StatusUnauthorized)
			return
		}

		allowed := limiter.Allow(apiKey)

		log.Printf("apiKey=%s allowed=%v\n", apiKey, allowed)

		if !allowed {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		fmt.Fprintln(w, "request allowed")
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
