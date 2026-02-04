# Rate-Limited HTTP API

## Overview
This project implements a production-style HTTP API that enforces rate limits under concurrent access.  
The focus is not on feature completeness, but on correctness, concurrency safety, and clear architectural decisions.

## Purpose & Motivation
Rate limiting is a common requirement in real-world backend systems, yet it is often treated as a black box or delegated entirely to third-party tools.

The goal of this project is to explore how rate limiting works internally by building it from scratch, reasoning about tradeoffs, and understanding the implications of different approaches under load and concurrent usage.

This project intentionally prioritizes clarity and correctness over premature optimization.

## High-Level Design
At a high level, the system consists of:
- An HTTP API responsible for handling client requests
- A rate limiter component that decides whether a request should be allowed or rejected
- A storage layer used to track request counts and time windows

The API identifies clients and applies rate limiting rules before processing requests.

## Design Decisions & Tradeoffs
Some of the decisions explored in this project include:
- Which rate limiting strategy to use (fixed window, sliding window, token bucket)
- How to safely manage shared state under concurrent access
- Tradeoffs between accuracy, fairness, and performance

Each approach is documented with its limitations and the reasons behind choosing or evolving it.

## Current Status
The project is currently in its early stages.  
Initial work focuses on defining the API contract and implementing a basic in-memory rate limiting strategy.


## Endpoints
### GET `/health` to check the health of the server
### GET `/protected` route to where the requests should be made to test the Rate Limit

## Getting Started

### Prerequisites
- Go 1.22 or newer

### Running the server

Clone the repository and run:
```bash
go run .
```

The server will start on: `http://localhost:8080`

### Testing protected endpoint

Send request with an API Key: `curl -H "X-API-Key: test" http://localhost:8080/protected`

To test the limits run:
```bash
for i in {1..12}; do
  curl -H "X-API-Key: test" localhost:8080/protected
done
```

After exceeding the configured limit, the server will respond with: `429 Too Many Requests`

## Roadmap
- Define the HTTP API contract
- Implement a fixed window rate limiter
- Add concurrency-safe storage
- Introduce improved rate limiting strategies
- Add observability and tests
