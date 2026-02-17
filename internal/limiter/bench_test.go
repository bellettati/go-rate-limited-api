package limiter

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkFixedWindow_Allow_SameKey_Parallel(b *testing.B) {
	clock := RealClock{}
	rl := NewFixedWindowLimiter(
		clock,
		LimitConfig{ Limit: 1_000_000_000, Window: time.Second},
		nil,
	)

	key := "same-key"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rl.Allow(key)
		}
	})
}

func BenchmarkFixedWindow_Allow_ManyKeys_Parallel(b *testing.B) {
	clock := RealClock{}
	rl := NewFixedWindowLimiter(
		clock,
		LimitConfig{Limit: 1_000_000_000, Window: time.Second},
		nil,
	)

	var counter uint64
	var mu sync.Mutex

	nextKey := func() string {
		mu.Lock()
		counter++
		v := counter
		mu.Unlock()
		return fmt.Sprintf("key-%d", v)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		key := nextKey()
		for pb.Next() {
			_ = rl.Allow(key)
		}
	})
}

func BenchmarkFixedWindow_Allow_Blocked_SameKey_Parallel(b *testing.B) {
	clock := RealClock{}
	rl := NewFixedWindowLimiter(clock, LimitConfig{Limit: 1, Window: time.Hour}, nil)

	key := "blocked-key"

	_ = rl.Allow(key)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rl.Allow(key)
		}
	})
}

func BenchmarkFixedWindow_Allow_FirstHit_UniqueKey(b *testing.B) {
	clock := RealClock{}
	rl := NewFixedWindowLimiter(clock, LimitConfig{Limit: 1_000_000_000, Window: time.Second}, nil)

	var seq uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := atomic.AddUint64(&seq, 1)
		key := fmt.Sprintf("key-%d", n)
		_ = rl.Allow(key)
	}
}

func BenchmarkSlidingWindow_Allow_Blocked_SameKey_Parallel(b *testing.B) {
	clock := RealClock{}
	rl := NewSlidingWindowLimiter(clock, LimitConfig{Limit: 1, Window: time.Hour}, nil)

	key := "blocked-key"
	_ = rl.Allow(key) // prime so subsequent calls are blocked

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rl.Allow(key)
		}
	})
}

func BenchmarkSlidingWindow_Allow_FirstHit_UniqueKey(b *testing.B) {
	clock := RealClock{}
	rl := NewSlidingWindowLimiter(clock, LimitConfig{Limit: 1_000_000_000, Window: time.Second}, nil)

	var seq uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := atomic.AddUint64(&seq, 1)
		key := fmt.Sprintf("key-%d", n)
		_ = rl.Allow(key)
	}
}

func BenchmarkTokenBucket_Allow_Blocked_SameKey_Parallel(b *testing.B) {
	clock := RealClock{}
	rl := NewTokenBucketLimiter(clock, LimitConfig{Limit: 1, Window: time.Hour}, nil)

	key := "blocked-key"
	_ = rl.Allow(key) // drain the only token

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rl.Allow(key)
		}
	})
}

func BenchmarkTokenBucket_Allow_FirstHit_UniqueKey(b *testing.B) {
	clock := RealClock{}
	rl := NewTokenBucketLimiter(clock, LimitConfig{Limit: 1_000_000_000, Window: time.Second}, nil)

	var seq uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := atomic.AddUint64(&seq, 1)
		key := fmt.Sprintf("key-%d", n)
		_ = rl.Allow(key)
	}
}
