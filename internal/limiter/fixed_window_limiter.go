package limiter

import (
	"context"
	"time"

	"github.com/bellettati/go-rate-limited-api/internal/store"
)

type FixedWindowLimiter struct {
	st 			 store.Store
	defaultLimit LimitConfig
	overrides    map[string]LimitConfig
	clock 		 Clock
}

func NewFixedWindowLimiter(st store.Store, clock Clock, defaultLimit LimitConfig, overrides map[string]LimitConfig) *FixedWindowLimiter {
	if overrides == nil {
		overrides = make(map[string]LimitConfig)
	}

	return &FixedWindowLimiter{
		st:		      st,
		defaultLimit: defaultLimit,
		overrides:    overrides,
		clock:        clock,
	}
}

func (rl *FixedWindowLimiter) configFor(apiKey string) LimitConfig {
	if cfg, ok := rl.overrides[apiKey]; ok {
		return cfg
	}

	return rl.defaultLimit
}

func windowBounds(now time.Time, window time.Duration) (start time.Time, end time.Time) {
	start = now.Truncate(window)
	end = start.Add(window)
	return start, end
}

func (rl *FixedWindowLimiter) Allow(apiKey string) RateLimitResult {
	cfg := rl.configFor(apiKey)

	now := rl.clock.Now()
	windowStart, windowEnd := windowBounds(now, cfg.Window)

	key := "rl:fixed:"+apiKey+":"+formatUnixNano(windowStart)

	ttl := time.Until(windowEnd)
	if ttl < 0 {
		ttl = 0
	}

	val, _, err := rl.st.IncrWithTTL(context.Background(), key, ttl)
	if err != nil {
		return RateLimitResult {
			Allowed:   true,
			Remaining: cfg.Limit,
			ResetAt:   windowEnd,
			Limit:     cfg.Limit,
		}
	}

	allowed := int(val) <= cfg.Limit
	remaining := cfg.Limit - int(val)

	if remaining < 0 {
		remaining = 0
	}

	return RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   windowEnd,
		Limit:     cfg.Limit,
	}
}

func formatUnixNano(t time.Time) string {
	var b [32]byte
	n := t.UnixNano()

	i := len(b)
	if n == 0 {
		i--
		b[i] = '0'
		return string(b[i:])
	}

	neg := n < 0
	if neg {
		n = -n
	}

	for n > 0 {
		i--
		b[i] = byte('0' + (n % 10))
		n /= 10
	}

	if neg {
		i--
		b[i] = '-'
	}

	return string(b[i:])
}