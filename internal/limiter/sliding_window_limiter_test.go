package limiter

import (
	"testing"
	"time"
)

func TestSlidingWindow_AllowWithinLimit(t *testing.T) {
	limit := 3
	window := 2 * time.Second

	clock := NewFakeClock(time.Now())
	limiter := NewSlidingWindowLimiter(
		clock,
		LimitConfig{Limit: limit, Window: window},
		nil,
	)

	apiKey := "test-api-key"

	for i := 0; i < limit; i++ {
		res := limiter.Allow(apiKey)

		if !res.Allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}

		expectedRemaining := limit - (i + 1)
		if res.Remaining != expectedRemaining {
			t.Fatalf("expected remaining %d, got %d", expectedRemaining, res.Remaining)
		}
	}
}

func TestSlidingWindow_RejectWhenLimitExceeds(t *testing.T) {
	limit := 2
	window := 2 * time.Second

	clock := NewFakeClock(time.Now())
	limiter := NewSlidingWindowLimiter(
		clock, 
		LimitConfig{Limit: limit, Window: window},
		nil,
	)

	apiKey := "test-api-key"

	for i := 0; i < limit; i++ {
		res := limiter.Allow(apiKey)
		if !res.Allowed {
			t.Fatalf("expected request to be allowed")
		}
	}

	res := limiter.Allow(apiKey)
	if res.Allowed {
		t.Fatalf("expected request to rejecte")
	}
}

func TestSlidingWindow_AllowedAfterWindowPasses(t *testing.T) {
	limit := 2
	window := 2 * time.Second

	clock := NewFakeClock(time.Now())
	limiter := NewSlidingWindowLimiter(
		clock,
		LimitConfig{Limit: limit, Window: window},
		nil,
	)

	apiKey := "api-test-key"


	for i := 0; i < limit; i++ {
		res := limiter.Allow(apiKey)
		if !res.Allowed {
			t.Fatalf("expected request to be allowed")
		}
	}

	res := limiter.Allow(apiKey)
	if res.Allowed {
		t.Fatalf("expected request to be rejected")
	}

	clock.Advance(window + 50 * time.Millisecond)

	res = limiter.Allow(apiKey)
	if !res.Allowed {
		t.Fatalf("expected request to be allowed")
	}
}

func TestSlidingWindow_IsPerClient(t *testing.T) {
	limit := 1
	window := 2 * time.Second

	clock := NewFakeClock(time.Now())
	limiter := NewSlidingWindowLimiter(
		clock,
		LimitConfig{Limit: limit, Window: window},
		nil,
	)

	apiKey1 := "test-one"
	apiKey2 := "test-two"

	res := limiter.Allow(apiKey1)
	if !res.Allowed {
		t.Fatalf("expected apiKey1 first request to be allowed")
	}
	
	res = limiter.Allow(apiKey1)
	if res.Allowed {
		t.Fatalf("expected apiKey1 first request to be rejected")
	}

	res = limiter.Allow(apiKey2)
	if !res.Allowed {
		t.Fatalf("expected apiKey2 first request to be allowed independently")
	}
}

func TestCleanUp_RemovesInactiveClients_SlidingWindow(t *testing.T) {
	clock := NewFakeClock(time.Now())

	sw := NewSlidingWindowLimiter(
		clock,
		LimitConfig{Limit: 5, Window: time.Minute},
		nil,
	)

	key := "test-key"

	sw.Allow(key)

	clock.Advance(5 * time.Minute);

	sw.cleanup()

	if _, exists := sw.clients[key]; exists {
		t.Fatalf("expected client to be removed after incativity")
	}
}