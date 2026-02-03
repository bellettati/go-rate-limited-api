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

## How to Run
Instructions will be added as the implementation progresses.

## Roadmap
- Define the HTTP API contract
- Implement a fixed window rate limiter
- Add concurrency-safe storage
- Introduce improved rate limiting strategies
- Add observability and tests
