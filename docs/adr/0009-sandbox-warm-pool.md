# 0009. Pre-warm a pool of idle sandbox containers per template

Date: 2026-07-15

## Status

Accepted

## Context

Every `CreateSandbox` call previously paid the full synchronous cost of `ImagePull` + `ContainerCreate` + `ContainerStart` on the request path. For any template whose image Docker hadn't already cached locally, this could add multiple seconds of latency to sandbox creation — a poor experience for what should feel like a near-instant operation, and the core "cold start" problem this system needs to solve to be usable.

Two general approaches exist for reducing this: make the cold path itself faster (smaller base images, layer caching tricks), or avoid paying the cold path's cost on the request path at all by keeping spare capacity ready ahead of time. The latter is what E2B itself does in production, and is more directly effective — no amount of image optimization fully eliminates container creation latency, but having an already-running container ready to hand out reduces perceived latency to nearly zero for the common case.

## Decision

Maintain a per-template pool of N idle, already-running containers (`sleep infinity`), configurable via `WARM_POOL_SIZE`. `CreateSandbox` attempts to take a container from the pool first; only falls back to a synchronous cold create if the pool is empty for that template. A background goroutine per template refills the pool both reactively (signaled immediately after a take) and on a periodic safety-net ticker, and verifies a container's liveness before handing it out (discarding any that died unexpectedly rather than returning a broken container ID to the caller).

## Consequences

- The common case (pool has capacity) returns a ready container near-instantly instead of paying pull/create latency; an `X-Sandbox-Warm-Start` response header makes this observable per-request for debugging and future measurement.
- Idle warm containers consume real host resources (memory, and CPU/disk to a lesser extent) continuously, regardless of actual demand — this is a deliberate trade of steady-state resource cost for lower request latency, the same fundamental tradeoff as Option A pause/resume (ADR 0005) but applied to the creation path instead of the pause path.
- There is currently no global cap across templates — total idle capacity scales linearly with `(number of templates × WARM_POOL_SIZE)`, with no upper bound on total resource usage across the whole pool. Adding many templates without adjusting `WARM_POOL_SIZE` down, or adding a global ceiling, risks unbounded idle resource consumption. This is flagged as necessary follow-up before adding templates freely in a real deployment.
- Pool sizing is currently static and uniform per template; it does not yet adapt to observed demand (e.g. a heavily-used template running out of warm capacity faster than a rarely-used one with the same pool size). Dynamic/adaptive sizing is a reasonable future enhancement once real usage data exists.