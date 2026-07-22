# 0008. Token bucket rate limiting via atomic Redis Lua script, failing open

Date: 2026-07-15

## Status

Accepted

## Context

Without any rate limiting, a single API key (buggy client, retry loop, or intentional abuse) could create unbounded sandboxes or exec calls, exhausting Docker/DB/host resources shared by every other user of the system.

Several algorithms were available: fixed window counters (simple, but allow up to 2x the intended rate at window boundaries), sliding window logs (accurate, but expensive to store per-request timestamps), and token bucket (allows legitimate bursts while enforcing a steady-state average rate, with cheap constant-size state per key).

Because Cage may run multiple replicas in the future (see ADR 0009), the rate limiter's state needed to be shared across processes — an in-memory limiter would allow each replica to independently grant a client its own full quota, defeating the purpose. Redis was the natural shared store, already present for caching (ADR 0007).

A naive "read tokens, check, write tokens" sequence against Redis has a race condition under concurrent requests from the same key: two requests could both read the same token count before either writes back the decrement, incorrectly allowing both through. This required the check-and-decrement to run as a single atomic operation.

## Decision

Implement a token bucket rate limiter as a Lua script executed atomically via Redis `EVAL`, keyed per API key hash. Following the same principle as ADR 0007, if Redis is unavailable, requests are allowed through (fail open) rather than blocked — rate limiting is a protective measure, not a correctness guarantee, and should not become a single point of failure for the entire API.

## Consequences

- Rate limiting is correct under concurrency and shared correctly across any number of future replicas, since all state lives in Redis and the check-and-decrement is atomic.
- A Redis outage means rate limiting is temporarily unenforced rather than the API becoming entirely unavailable — an intentional tradeoff prioritizing availability of the core product over strict enforcement of a protective mechanism.
- The current implementation applies one global bucket per API key across all routes. Sandbox creation (expensive: Docker pull/create) and read operations like `GET /sandboxes/{id}` (cheap) currently share the same quota, meaning a client's read traffic can be throttled by their own write traffic and vice versa. Splitting into per-route or per-operation-class buckets is a known follow-up once real usage patterns are available to tune limits against.