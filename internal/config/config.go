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

type RateLimitBackend string

const (
	InMemory RateLimitBackend = "in_memory"
	Redis RateLimitBackend = "redis"
)

type Config struct {
	RateLimitStrategy RateLimitStrategy 
	RateLimitBackend RateLimitBackend

	DefaultLimit      int
	DefaultWindow     time.Duration

	RedisAddr string
	RedisPassword string
	RedisDB int
	RedisDialTimeout time.Duration
	RedisReadTimeout time.Duration
	RedisWriteTimeout time.Duration
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

func getEnvAsDurationSeconds(key string, defaultSeconds int) time.Duration {
	secs := getEnvAsInt(key, defaultSeconds)
	if secs <= 0 {
		return time.Duration(defaultSeconds) * time.Second
	}
	return time.Duration(secs) * time.Second
}

func normalizeStrategy(s string) RateLimitStrategy {
	return RateLimitStrategy(strings.ToLower(strings.TrimSpace(s)))
}

func normalizeBackend(s string) RateLimitBackend {
	return RateLimitBackend(strings.ToLower(strings.TrimSpace(s)))
}

func validateStrategy(s RateLimitStrategy) bool {
	switch s {
	case FixedWindow, SlidingWindow, TokenBucket:
		return true
	default:
		return false
	}
}

func validateBackend(b RateLimitBackend) bool {
	switch b {
	case InMemory, Redis:
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

	rawBackend := getEnv("RATE_LIMIT_BACKEND", string(InMemory))
	backend := normalizeBackend(rawBackend)
	if !validateBackend(backend) {
		log.Fatalf(
			"Invalid RATE_LIMIT_BACKEND=%q (expected: %s, %s)",
			rawBackend,
			InMemory,
			Redis,
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

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvAsInt("REDIS_DB", 0)

	redisDialTimeout := getEnvAsDurationSeconds("REDIS_DIAL_TIMEOUT_SECONDS", 2)
	redisReadTimeout := getEnvAsDurationSeconds("REDIS_READ_TIMEOUT_SECONDS", 2)
	redisWriteTimeout:= getEnvAsDurationSeconds("REDIS_WRITE_TIMEOUT_SECONDS", 2)

	return Config{
		RateLimitStrategy: strategy,
		RateLimitBackend: backend,

		DefaultLimit: limit,
		DefaultWindow: time.Duration(windowSeconds) * time.Second,

		RedisAddr: redisAddr,
		RedisPassword: redisPassword,
		RedisDB: redisDB,
		RedisDialTimeout: redisDialTimeout,
		RedisReadTimeout: redisReadTimeout,
		RedisWriteTimeout: redisWriteTimeout,
	}
}
