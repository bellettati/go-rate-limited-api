package main

import (
	"sync"
	"testing"
	"time"
)

func TestAllow_FirstRequestAllowed(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	result := rl.Allow("test-key")

	if !result.Allowed {
		t.Fatalf("expected first request to be allowed")
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	key := "test-key"

	rl.Allow(key)
	rl.Allow(key)

	result := rl.Allow(key)

	if result.Allowed {
		t.Fatalf("expected request to be blocked after limit exceeded")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	key := "test-key"

	first := rl.Allow(key)
	if !first.Allowed {
		t.Fatalf("expected first request to be allowed")
	}

	second := rl.Allow(key)
	if second.Allowed {
		t.Fatalf("expected second request to be blocked")
	}

	time.Sleep(60 * time.Millisecond)

	third := rl.Allow(key)
	if !third.Allowed {
		t.Fatalf("expected third request to be allowed")
	}
}

func TestAllow_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, time.Millisecond)

	key := "test-key"

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Allow(key)
		}()
	}

	wg.Wait()
}
