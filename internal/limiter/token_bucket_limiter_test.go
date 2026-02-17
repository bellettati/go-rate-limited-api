package limiter

import (
	"testing"
	"time"
)

func TestTokenBucketAllowsInitialBurst(t *testing.T) {
	clock := NewFakeClock(time.Now())
	limiter := NewTokenBucketLimiter(
		clock,
		LimitConfig{Limit: 2, Window: time.Minute},
		nil,
	)

	res1 := limiter.Allow("test-key")
	res2 := limiter.Allow("test-key")
	res3 := limiter.Allow("test-key")

	if !res1.Allowed || !res2.Allowed {
		t.Fatalf("expected first two requests allowed")
	}

	if res3.Allowed {
		t.Fatalf("expected third request to be denied")
	}
}

func TestTokenBucketRefillsOverTime(t *testing.T) {
	clock := NewFakeClock(time.Now())
	limiter := NewTokenBucketLimiter(
		clock,
		LimitConfig{Limit: 2, Window: time.Second},
		nil,
	)

	limiter.Allow("test-key")
	clock.Advance(time.Second)
	res := limiter.Allow("test-key")

	if !res.Allowed {
		t.Fatalf("expected token to refill")
	}
}

func TestTokenBucketPartialRefill(t *testing.T) {
	clock := NewFakeClock(time.Now())
	limiter := NewTokenBucketLimiter(
		clock,
		LimitConfig{Limit: 10, Window: 10 * time.Second},
		nil,
	)

	for i := 0; i < 10; i++ {
		limiter.Allow("test-key")
	}

	clock.Advance(time.Second)

	res := limiter.Allow("test-key")
	if !res.Allowed {
		t.Fatalf("expected partial refill to allow request")
	}
}

func TestTokenBucketDoesNotOverfill(t *testing.T) {
	clock := NewFakeClock(time.Now())
	limiter := NewTokenBucketLimiter(
		clock,
		LimitConfig{Limit: 2, Window: time.Second},
		nil,
	)

	clock.Advance(5 * time.Second)

	res1 := limiter.Allow("test-key")
	res2 := limiter.Allow("test-key")
	res3 := limiter.Allow("test-key")

	if !res1.Allowed || !res2.Allowed {
		t.Fatalf("expected full bucket")
	}

	if res3.Allowed {
		t.Fatalf("bucket should not exceed capacity")
	}
}

func TestTokenBucketIsolatedPerKey(t *testing.T) {
	clock := NewFakeClock(time.Now())
	limiter := NewTokenBucketLimiter(
		clock,
		LimitConfig{Limit: 1, Window: time.Minute},
		nil,
	)

	res1 := limiter.Allow("test-key")
	res2 := limiter.Allow("test-key-2")

	if !res1.Allowed || !res2.Allowed {
		t.Fatalf("keys should have indpendent buckets")
	}
}

func TestTokenBucketDeniesWhenEmpty(t *testing.T) {
	clock := NewFakeClock(time.Now())
	limiter := NewTokenBucketLimiter(
		clock,
		LimitConfig{Limit: 1, Window: time.Minute},
		nil,
	)

	limiter.Allow("test-key")
	res := limiter.Allow("test-key")

	if res.Allowed {
		t.Fatalf("expected denial when bucket is empty")
	}
}

func TestCleanUp_RemovesInactiveClients_TokenBucket(t *testing.T) {
	clock := NewFakeClock(time.Now())

	tb := NewTokenBucketLimiter(
		clock,
		LimitConfig{Limit: 5, Window: time.Minute},
		nil,
	)

	key := "test-key"

	tb.Allow(key)

	clock.Advance(5 * time.Minute);

	tb.cleanup()

	if _, exists := tb.clients[key]; exists {
		t.Fatalf("expected client to be removed after incativity")
	}
}