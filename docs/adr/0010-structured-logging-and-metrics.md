# 0010. Structured logging via slog and metrics via Prometheus

Date: 2026-07-16

## Status

Accepted

## Context

Logging up to this point was a mix of `log.Printf`/`log.Println`/`log.Fatal` calls with free-text, string-interpolated messages (e.g. `"reaper: killing expired sandbox %s"`). This is adequate for local development but becomes a real liability once the service runs anywhere it can't simply be watched in a terminal — free-text logs can't be reliably filtered by field (e.g. "show me every log line for sandbox X" requires fragile string matching, not a structured field lookup), and most log aggregation platforms parse structured JSON far more usefully than plain text.

Separately, the service had no way to answer basic operational questions without reading logs line-by-line: request rate, latency distribution, error rate, warm pool hit/miss ratio, how many sandboxes are currently active. These are exactly the kind of aggregate, time-series questions logs are poorly suited to answer but a metrics system is designed for.

Go's standard library has included `log/slog` (structured logging) since 1.21, avoiding a third-party logging dependency. Prometheus is the de facto standard for pull-based metrics in the Go ecosystem, with first-class client libraries and broad compatibility with free-tier observability tooling.

## Decision

Adopt `slog` with a JSON handler as the sole logging mechanism, replacing all `log.Printf`/`log.Fatal` calls. Adopt Prometheus client libraries for metrics, exposing a `/metrics` endpoint (unauthenticated, matching Prometheus scrape conventions) with counters and histograms for HTTP requests, warm pool hits/misses, and sandboxes reaped.

## Consequences

- Every request is tagged with a correlation ID (via chi's `middleware.RequestID`) and logged with structured fields (method, path, status, duration) in a single line per request — this makes tracing a single request's lifecycle, or filtering by any field, straightforward once logs are shipped to any aggregation platform.
- Prometheus metrics use route *patterns* (e.g. `/sandboxes/{id}`), not raw request paths with real sandbox IDs substituted in — this was a deliberate choice to avoid unbounded metric cardinality, a common and costly mistake in Prometheus instrumentation where using raw dynamic values as label values causes the number of unique time series to grow without bound.
- The `/metrics` endpoint is currently public/unauthenticated, consistent with standard Prometheus scraping conventions, but this does expose operational data (request volumes, active sandbox counts) to anyone who can reach the service. Acceptable for the current stage; worth restricting to an internal network or adding separate auth before wider exposure.
- Structured logging and metrics are complementary, not redundant: metrics answer aggregate "how much/how often" questions cheaply at query time, logs answer "what exactly happened in this one case" — both were adopted rather than treating one as a substitute for the other.