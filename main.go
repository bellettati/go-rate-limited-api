package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	defaultLimit := LimitConfig{
		Limit:  10,
		Window: time.Minute,
	}

	overrides := map[string]LimitConfig{
		"vip": LimitConfig{Limit: 3, Window: time.Minute},
	}

	rl := NewRateLimiter(defaultLimit, overrides)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

	rateLimitedMux := RateLimit(rl)(mux)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", rateLimitedMux))
}
