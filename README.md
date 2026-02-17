# Rate-Limited HTTP API

## Overview
This project implements a production-style HTTP API with built-in rate limiting designed to behave correctly under concurrent access.

The focus is not on feature completeness, but on **correctness, concurrency safety, performance characteristics, and explicit architectural tradeoffs**, closely mirroring real-world backend systems.

---

## Purpose & Motivation
Rate limiting is a common requirement in backend systems, yet it is often abstracted away behind proxies or third-party services.

The goal of this project is to:

- Understand how rate limiting works internally
- Explore tradeoffs between different rate limiting strategies
- Build a correct, testable, and benchmarked implementation under concurrent load
- Design the system so it can evolve into a distributed limiter

This project intentionally prioritizes **clarity, correctness, and observability over premature optimization**.

---

## High-Level Design
At a high level, the system consists of:

- An HTTP server responsible for request handling
- A pluggable rate limiter component that decides whether a request is allowed
- An in-memory storage layer tracking request state per API key
- Middleware enforcing limits before request processing

Clients are identified via an API key sent in request headers.

---

## Rate Limiting Strategies

The limiter is configurable via environment variables and supports multiple algorithms:

| Strategy | Description | Characteristics |
|--------|-------------|----------------|
| Fixed Window | Counts requests in fixed intervals | Simple, predictable, fast |
| Sliding Window | Tracks timestamps per request | More accurate fairness |
| Token Bucket | Refill-based token model | Smooth rate limiting |

The strategy can be selected without code changes:

RATE_LIMIT_STRATEGY=fixed_window
RATE_LIMIT_STRATEGY=sliding_window
RATE_LIMIT_STRATEGY=token_bucket


---

## Architecture Design Decisions & Tradeoffs

### Pluggable Strategy Interface
All limiters implement a shared interface, allowing runtime selection of strategy without modifying application code.

This enables:
- clean separation of algorithm vs application logic
- easy experimentation with strategies
- future distributed implementations

---

### In-Memory Storage
The limiter stores per-key state in memory.

**Advantages**
- simple
- fast
- easy to reason about

**Tradeoffs**
- limits reset on restart
- does not scale across instances

The architecture is intentionally structured so storage can later be replaced with Redis or another distributed store.

---

### Concurrency Safety
Shared state is protected with mutexes to ensure correctness under concurrent access.

The implementation prioritizes:
- determinism
- correctness
- predictability

over premature lock-free complexity.

Concurrency behavior is validated via parallel test cases.

---

### Automatic Memory Cleanup
Inactive clients are periodically removed from memory.

Without cleanup, high-cardinality keys could cause unbounded memory growth.  
Cleanup ensures the limiter remains stable for long-running processes.

---

### Deterministic Time via Clock Abstraction
The limiter does not directly depend on system time.

Instead, it receives a `Clock` interface:

- `RealClock` → production
- `FakeClock` → tests

This allows:

- deterministic tests
- no sleeps in test suite
- precise control over time-dependent behavior

---

### Middleware-Based Enforcement
Rate limiting is implemented as HTTP middleware.

This cleanly separates:

- request handling
- rate limiting logic
- application endpoints

This mirrors real production backend architectures.

---

## Performance & Benchmarking

Benchmarks are included to measure performance of:

- steady-state request path
- blocked request path
- first request for new client
- parallel contention behavior

**Representative results (Apple M4 Pro):**

| Scenario | Result |
|--------|-------|
| Hot path | ~130–145 ns/op |
| Allocations | 0 allocs/op |
| Blocked path | same speed as allowed |
| First request | small bounded allocations |

The system is optimized so **steady-state requests produce zero heap allocations**, minimizing GC pressure.

Benchmarks help validate not just speed, but **memory safety and algorithmic behavior**.

---

## Testing Strategy
The project includes automated tests focused on correctness and concurrency:

- Unit tests validating limiter behavior
- Concurrency tests simulating parallel requests
- Deterministic time tests using FakeClock
- Cleanup behavior validation

Tests validate system behavior, not just individual functions.

---

## Project Structure
The repository follows a production-style Go layout:

cmd/server → application entrypoint
internal/limiter → rate limiting algorithms
internal/middleware → HTTP middleware
internal/config → environment configuration
internal/handlers → endpoints


This structure keeps domain logic isolated and makes the system easier to extend.

---

## Current Capabilities
The system currently supports:

- Multiple rate limiting strategies
- Runtime strategy selection via config
- Concurrency-safe state handling
- Automatic cleanup of inactive clients
- Deterministic testing via injected clock
- Benchmarks for performance analysis
- Production-style project structure

---

## Endpoints

### `GET /health`
Health check endpoint.

---

### `GET /protected`
Protected endpoint subject to rate limiting.

---

## Getting Started

### Prerequisites
- Go 1.22+

---

### Run Without Building
go run ./cmd/server

---

### Build Binary
go build -o server ./cmd/server
./server

---

### Example `.env`
RATE_LIMIT_STRATEGY=token_bucket
DEFAULT_LIMIT=10
DEFAULT_WINDOW_SECONDS=60

---

## Future Extensions
Planned improvements include:

- Redis-backed distributed limiter
- Horizontal scaling support
- Metrics integration
- Structured logging
- Adaptive rate limits
- Per-endpoint limits

---

## Summary
This project is not just a rate limiter implementation — it is an exploration of backend engineering concerns:

- concurrency correctness
- performance characteristics
- testability
- architecture flexibility

It is designed as a foundation for building production-grade rate limiting systems.
