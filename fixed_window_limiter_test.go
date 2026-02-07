package main

import (
	"sync"
	"testing"
	"time"
)

func TestAllow_FirstRequestAllowed(t *testing.T) {
	rl := NewFixedWindowLimiter(LimitConfig{Limit: 5, Window: time.Minute}, nil)

	result := rl.Allow("test-key")

	if !result.Allowed {
		t.Fatalf("expected first request to be allowed")
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	rl := NewFixedWindowLimiter(LimitConfig{Limit: 2, Window: time.Minute}, nil)

	key := "test-key"

	rl.Allow(key)
	rl.Allow(key)

	result := rl.Allow(key)

	if result.Allowed {
		t.Fatalf("expected request to be blocked after limit exceeded")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	rl := NewFixedWindowLimiter(LimitConfig{Limit: 1, Window: 50 * time.Millisecond}, nil)

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
	rl := NewFixedWindowLimiter(LimitConfig{Limit: 100, Window: time.Millisecond}, nil)

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

func TestAllow_OverrideTakesPrecedence(t *testing.T) {
	rl := NewFixedWindowLimiter(
		LimitConfig{Limit: 5, Window: time.Minute},
		map[string]LimitConfig{
			"vip": {Limit: 1, Window: time.Minute},
		},
	)

	rl.Allow("vip")
	result := rl.Allow("vip")

	if result.Allowed {
		t.Fatalf("expected override limit to be enforced")
	}
}
