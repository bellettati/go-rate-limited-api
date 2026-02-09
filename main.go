package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := LoadConfig()

	defaultLimit := LimitConfig{
		Limit:  cfg.DefaultLimit,
		Window: cfg.DefaultWindow,
	}

	overrides := map[string]LimitConfig{
		"vip": {Limit: 3, Window: time.Minute},
	}

	var limiter Limiter

	switch cfg.RateLimitStrategy {
	case "token_bucket":
		limiter = NewTokenBucketLimiter(defaultLimit, overrides)
	default:
		limiter = NewFixedWindowLimiter(defaultLimit, overrides)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

	rateLimitedMux := RateLimit(limiter)(mux)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", rateLimitedMux))
}
