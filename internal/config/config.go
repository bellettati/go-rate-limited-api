package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type RateLimitStrategy string

const (
	FixedWindow RateLimitStrategy = "fixed_window"
	SlidingWindow RateLimitStrategy = "sliding_window"
	TokenBucket RateLimitStrategy = "token_bucket"
)

type Config struct {
	RateLimitStrategy RateLimitStrategy 
	DefaultLimit      int
	DefaultWindow     time.Duration
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}

	return val
}

func normalizeStrategy(s string) RateLimitStrategy {
	return RateLimitStrategy(strings.ToLower(strings.TrimSpace(s)))
}

func validateStrategy(s RateLimitStrategy) bool {
	switch s {
	case FixedWindow, SlidingWindow, TokenBucket:
		return true
	default:
		return false
	}
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using syatem env vars")
	}

	rawStrategy := getEnv("RATE_LIMIT_STRATEGY", string(FixedWindow))
	strategy := normalizeStrategy(rawStrategy)
	if !validateStrategy(strategy) {
		log.Fatalf(
			"Invalid RATE_LIMIT_STRATEGY=%q (expected: %s, %s, %s)",
			rawStrategy, FixedWindow, SlidingWindow, TokenBucket,
		)
	}

	limit := getEnvAsInt("DEFAULT_LIMIT", 10)
	windowSeconds := getEnvAsInt("DEFAULT_WINDOW_SECONDS", 60)

	if limit <= 0 {
		log.Fatalf("DEFAULT_LIMIT must be > 0 (got %d)", limit)
	}
	if windowSeconds <= 0 {
		log.Fatalf("DEFAULT_WINDOW_SECONDS must be > 0 (got %d)", limit)
	}

	return Config{
		RateLimitStrategy: strategy,
		DefaultLimit: limit,
		DefaultWindow: time.Duration(windowSeconds) * time.Second,
	}
}
