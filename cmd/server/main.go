package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bellettati/go-rate-limited-api/internal/config"
	"github.com/bellettati/go-rate-limited-api/internal/handlers"
	"github.com/bellettati/go-rate-limited-api/internal/limiter"
	"github.com/bellettati/go-rate-limited-api/internal/middleware"
)

func main() {
	cfg := config.LoadConfig()

	defaultLimit := limiter.LimitConfig{
		Limit:  cfg.DefaultLimit,
		Window: cfg.DefaultWindow,
	}

	overrides := map[string]limiter.LimitConfig{
		"vip": {Limit: 3, Window: time.Minute},
	}

	var requestLimiter limiter.Limiter

	switch cfg.RateLimitStrategy {
	case "token_bucket":
		requestLimiter = limiter.NewTokenBucketLimiter(defaultLimit, overrides)
	default:
		requestLimiter = limiter.NewFixedWindowLimiter(defaultLimit, overrides)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/protected", handlers.Protected)

	rateLimitedMux := middleware.RateLimit(requestLimiter)(mux)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", rateLimitedMux))
}
