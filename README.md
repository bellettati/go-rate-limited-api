# Rate-Limited HTTP API

## Overview
This project implements a production-style HTTP API with built-in rate limiting designed to behave correctly under concurrent access.

The focus is not on feature completeness, but on **correctness, concurrency safety, observability, and explicit architectural tradeoffs**, closely mirroring real-world backend constraints.

---

## Purpose & Motivation
Rate limiting is a common requirement in backend systems, yet it is often abstracted away behind proxies or third-party services.

The goal of this project is to:
- Understand how rate limiting works internally
- Explore the tradeoffs of different strategies
- Build a correct and testable implementation under concurrent load

This project intentionally prioritizes **clarity and correctness over premature optimization**.

---

## High-Level Design
At a high level, the system consists of:

- An HTTP server responsible for request handling
- A rate limiter component that decides whether a request is allowed
- An in-memory storage layer tracking request counts per API key
- Middleware enforcing rate limits before request processing

Clients are identified via an API key sent in request headers.

---

## Design Decisions & Tradeoffs

### Rate Limiting Strategy
- **Fixed window rate limiting** was chosen for simplicity and clarity.
- While not perfectly fair at window boundaries, it is easy to reason about and test.
- The design allows future extension to sliding window or token bucket strategies.

### Storage
- An **in-memory map** is used to track request counts.
- This keeps the implementation simple and focused on correctness.
- The tradeoff is that limits reset on process restart and do not scale across instances.

### Concurrency & Safety
- Shared state is protected using mutexes to ensure correctness under concurrent access.
- Concurrency behavior is explicitly tested using parallel goroutines.
- The implementation favors safety and predictability over lock-free complexity.

### Middleware-Based Enforcement
- Rate limiting is applied via HTTP middleware.
- This cleanly separates request handling from enforcement logic and mirrors real production setups.

---

## Logging & Observability
Basic structured logging is implemented to improve visibility into system behavior:
- Rate limit decisions (allowed / rejected)
- Request processing duration
- API keys are **masked** in logs to avoid leaking sensitive data

This provides insight into system behavior without introducing external observability dependencies.

---

## Testing Strategy
The project includes automated tests focused on correctness and concurrency:
- Unit tests validating rate limiter behavior
- Concurrency tests simulating multiple parallel requests using the same API key
- Tests ensure limits are enforced correctly under load

The goal is to validate behavior, not just individual functions.

---

## Current Status
Core functionality is implemented:
- Fixed window rate limiter
- Concurrency-safe in-memory storage
- HTTP middleware enforcement
- Basic logging
- Concurrent test coverage

The project is in a solid, functional state and designed to be incrementally extended.

---

## Endpoints

### `GET /health`
Health check endpoint.

### `GET /protected`
Protected endpoint subject to rate limiting.

---

## Getting Started

### Prerequisites
- Go 1.22 or newer

### Running the Server
```bash
go run .
