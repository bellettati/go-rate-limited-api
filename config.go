package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type RateLimitStrategy string

const (
	FixedWindow RateLimitStrategy = "fixed_window"
	TokenBucket RateLimitStrategy = "token_bucket"
)

type Config struct {
	RateLimitStrategy string
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

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using syatem env vars")
	}

	strategy := getEnv("RATE_LIMIT_STRATEGY", "fixed_window")
	limit := getEnvAsInt("DEFAULT_LIMIT", 10)
	windowSeconds := getEnvAsInt("DEFAULT_WINDOW_SECONDS", 60)

	return Config{
		RateLimitStrategy: strategy,
		DefaultLimit:      limit,
		DefaultWindow:     time.Duration(windowSeconds) * time.Second,
	}
}
