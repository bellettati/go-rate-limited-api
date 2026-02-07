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
- Low computational overhead
- Easy to reason about under concurrency

### Limitations
- **Burst problem at window boundaries**
  - A client can send 10 requests at the end of one window and 10 more at the start of the next
- Less fair under spiky traffic
- Not ideal for APIs exposed to unpredictable load

### When to Use
- Internal services
- Low-traffic APIs
- Systems where simplicity is more important than precision

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
- Smoother traffic control
- Supports bursty usage patterns
- More realistic model for real-world APIs

### Limitations
- Slightly more complex to implement
- Requires floating-point math and time-based calculations
- Harder to reason about exact limits compared to fixed windows

### When to Use
- Public APIs
- Systems with uneven or bursty traffic
- User-facing endpoints where fairness matters

---

## Comparison Summary

| Strategy      | Burst Handling | Fairness | Complexity | Typical Use Case |
|---------------|---------------|----------|------------|------------------|
| Fixed Window  | ❌ Poor        | Medium   | Low        | Simple/internal APIs |
| Token Bucket  | ✅ Good        | High     | Medium     | Public APIs, SaaS |

---

## Current Project Scope

In this project:
- Both **Fixed Window** and **Token Bucket** strategies are implemented
- State is kept **in-memory**
- Concurrency safety is enforced via mutexes
- Tests validate behavior under basic scenarios

This intentionally avoids:
- Distributed state (Redis, Memcached)
- Persistent storage
- Network-level rate limiting

Those concerns are deferred to later phases.

---

## Future Improvements

Planned evolutions include:
- Strategy selection via interface-based design
- External storage for distributed rate limiting
- Per-route and per-method limits
- Improved observability and metrics
- Better client identification mechanisms

Each improvement will be added only after clearly understanding the tradeoffs involved.

---

## Final Note

Rate limiting is not just an algorithm — it is a **design decision**.

This project treats rate limiting as a first-class concern, focusing on correctness, clarity, and real-world applicability rather than premature optimization.
