# 0007. Cache API key validation in Redis, failing open on cache errors

Date: 2026-07-15

## Status

Accepted

## Context

Every authenticated request validates its API key against Postgres via `ValidateAPIKey`, adding a DB round-trip to the hottest path in the system — this happens on every single `/sandboxes/*` request, regardless of which operation is being performed. API keys are read far more often than they change (created once, occasionally revoked), making this an ideal caching candidate.

Two questions needed answering: what to cache, and what should happen if the cache itself is unavailable.

On the first question: caching only successful validations was considered, but this leaves the DB exposed to repeated lookups from a misconfigured or malicious client retrying a bad key indefinitely — caching negative (`invalid`) results too protects against this at negligible cost.

On the second question, two options existed: **fail closed** (treat a Redis outage as "reject the request," since we can't confirm validity) or **fail open on the cache layer specifically** (fall through to querying Postgres directly, treating Redis purely as an optimization). Failing closed would mean a Redis outage — a component whose only job is to make things faster — could take down the entire API's ability to authenticate anyone, which is a disproportionate blast radius for what should be a non-critical-path dependency.

## Decision

Cache both valid and invalid API key hash lookups in Redis with a 5-minute TTL. If Redis is unreachable or errors, fall through to querying Postgres directly rather than rejecting the request — Redis is a performance optimization for the auth path, not a correctness dependency.

## Consequences

- Measured ~40% reduction in average request latency on the auth path in local testing (2.08ms → 1.23ms average over 50 sequential requests).
- A Redis outage degrades performance (every request falls back to Postgres) but does not cause an outage of the API itself — this tradeoff was deliberately chosen over stricter cache-must-be-available semantics.
- Revoking an API key does not take effect instantly for any client whose validation result is already cached — up to a 5-minute window where a revoked key could still be accepted, unless the revocation path explicitly deletes the corresponding cache entry (`apikey:<hash>`). Any future key-revocation endpoint must remember to invalidate this cache key, or the TTL becomes a real security gap rather than just a staleness inconvenience.
- The same fail-open principle was later applied to rate limiting (see ADR 0008) for consistency — protective/optimizing middleware degrades gracefully, core request handling does not.