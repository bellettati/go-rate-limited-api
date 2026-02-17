# Rate Limiting Strategies

This document describes the rate limiting strategies implemented in this project, the problems they solve, and the tradeoffs involved in each approach.

The goal is not to provide an exhaustive catalog of algorithms, but to deeply understand **why** and **when** each strategy should be used in real-world systems.

---

## Why Rate Limiting Matters

Rate limiting protects systems from:

- Abuse and denial-of-service scenarios  
- Accidental overload caused by buggy clients  
- Resource exhaustion (CPU, memory, database connections)

A well-designed rate limiter balances:

- Fairness  
- Accuracy  
- Performance  
- Simplicity  

There is no single “best” strategy — only strategies that fit different contexts.

---

## Implemented Strategies

This project currently implements three rate limiting strategies:

- Fixed Window
- Sliding Window
- Token Bucket

Each algorithm is exposed through a common interface and can be selected at runtime via configuration, allowing direct comparison of their behavior and tradeoffs.

---

## Fixed Window Rate Limiting

### How it Works
- Each client has a counter associated with a fixed time window
- Requests increment the counter
- Once the limit is reached, further requests are denied until the window resets

Example:
- Limit: 10 requests per minute
- Window: `12:00:00 → 12:00:59`
- Counter resets at `12:01:00`

### Advantages
- Simple to understand and implement
- Very fast execution path
- Minimal memory overhead
- Easy to reason about under concurrency

### Limitations
- **Burst problem at window boundaries**
  - A client can send 10 requests at the end of one window and 10 more at the start of the next
- Less fair under spiky traffic
- Can allow short bursts that exceed intended average rate

### When to Use
- Internal services
- Low-traffic APIs
- Systems where simplicity is more important than precision
- Situations where performance must be maximized

---

## Sliding Window Rate Limiting

### How it Works
- Each request timestamp is recorded per client
- On each request:
  - Old timestamps outside the window are removed
  - Remaining timestamps represent requests within the active window
- Requests are allowed only if the number of timestamps is below the limit

Example:
- Limit: 10 requests per minute
- Instead of resetting counters, the limiter continuously evaluates requests within the last 60 seconds

### Advantages
- More accurate than fixed windows
- Eliminates boundary burst problem
- Provides fairer distribution of requests over time
- Enforces true rolling limits

### Limitations
- Higher memory usage (timestamps per request)
- O(n) pruning cost per request (n = requests in window)
- More complex than fixed window
- Can become expensive for extremely high limits

### Implementation Note
The implementation caps internal slice preallocation to avoid excessive memory allocation for large limits. This ensures safety even if configuration values are very high.

### When to Use
- APIs requiring fairness
- Systems with spiky traffic
- Environments where accurate enforcement matters more than raw speed

---

## Token Bucket Rate Limiting

### How it Works
- Each client has a “bucket” of tokens
- Tokens are added gradually over time at a fixed refill rate
- Each request consumes one token
- Requests are denied when no tokens remain

Key characteristics:
- Allows short bursts
- Enforces an average request rate over time

### Advantages
- Smooth traffic shaping
- Supports bursty usage patterns
- More realistic model for real-world APIs
- Prevents sharp traffic spikes while allowing flexibility

### Limitations
- Slightly more complex implementation
- Requires time-based math calculations
- Slightly higher per-request computation than fixed window

### When to Use
- Public APIs
- User-facing systems
- Systems with bursty traffic patterns
- Services where fairness and smoothness matter

---

## Comparison Summary

| Strategy       | Burst Handling | Fairness | Accuracy | Complexity | Memory Use | Typical Use Case |
|----------------|---------------|----------|----------|------------|------------|------------------|
| Fixed Window   | ❌ Poor        | Medium   | Medium   | Low        | Low        | Internal APIs |
| Sliding Window | ✅ Excellent   | High     | High     | Medium     | Medium     | Fair usage enforcement |
| Token Bucket   | ✅ Good        | High     | Medium   | Medium     | Low        | Public APIs / SaaS |

---

## Runtime Strategy Selection

The limiter strategy can be selected at runtime using environment variables:

RATE_LIMIT_STRATEGY=fixed_window
RATE_LIMIT_STRATEGY=sliding_window
RATE_LIMIT_STRATEGY=token_bucket

This allows:

- Switching algorithms without recompilation
- Performance comparisons
- Behavioral testing
- Production tuning

---

## Current Project Scope

In this project:

- All strategies implement a shared limiter interface
- State is stored **in memory**
- Concurrency safety is enforced via mutexes
- Background cleanup removes inactive clients
- Benchmarks measure algorithm performance characteristics
- Time is injected via a Clock interface for deterministic testing

This intentionally avoids:

- Distributed state (Redis, Memcached)
- Persistent storage
- Network-level rate limiting

Those concerns are deferred to later stages.

---

## Performance Observations

Benchmarks show:

- Steady-state limiter checks execute in ~130–150 ns
- Zero heap allocations occur on the hot path
- First request for a new client incurs small bounded allocations
- All algorithms remain stable under parallel load

These measurements validate both correctness and efficiency.

---

## Future Improvements

Planned evolutions include:

- Distributed rate limiting (Redis-backed)
- Sharded or lock-optimized storage
- Adaptive rate limits
- Per-endpoint limits
- Dynamic configuration reload
- Metrics and observability integration

Each improvement will be added only after clearly understanding the tradeoffs involved.

---

## Final Note

Rate limiting is not just an algorithm — it is a **system design decision**.

Different strategies optimize for different constraints:

- simplicity
- fairness
- performance
- memory usage
- burst tolerance

This project treats rate limiting as a first-class architectural concern, focusing on correctness, clarity, and real-world applicability rather than premature optimization.
