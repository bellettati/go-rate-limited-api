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
	"github.com/bellettati/go-rate-limited-api/internal/store"
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
	clock := limiter.RealClock{} 

	var st store.Store	
	switch cfg.RateLimitBackend {
	case config.InMemory:
		st = store.NewMemoryStoreWithCleanupInterval(cfg.DefaultWindow)
	case config.Redis:
		rs, err := store.NewRedisStore(store.RedisConfig{
			Addr: cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB: cfg.RedisDB,
			DialTimeout: cfg.RedisDialTimeout,
			ReadTimeout: cfg.RedisReadTimeout,
			WriteTimeout: cfg.RedisWriteTimeout,
		})

		if err != nil {
			log.Fatal(err)
		}

		st = rs
	default:
		log.Fatalf("unsupported backend: %q", cfg.RateLimitBackend)
	}
	defer func() { _ = st.Close() }()

	switch cfg.RateLimitStrategy {
	case config.FixedWindow:
		requestLimiter = limiter.NewFixedWindowLimiter(st, clock, defaultLimit, overrides)
	case config.SlidingWindow:
		requestLimiter = limiter.NewSlidingWindowLimiter(clock, defaultLimit, overrides)
	case config.TokenBucket:
		requestLimiter = limiter.NewTokenBucketLimiter(clock, defaultLimit, overrides)
	default:
		log.Fatalf("unsupported rate limit strategy: %q", cfg.RateLimitStrategy)
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
