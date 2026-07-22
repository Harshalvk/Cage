# 0011. Graceful shutdown via signal-driven context cancellation

Date: 2026-07-16

## Status

Accepted

## Context

The service previously ran via a bare `http.ListenAndServe` call inside `main()`, with setup logic (config, DB, cache, background goroutines) also directly in `main()`. This had two related problems.

First, `main()` used `log.Fatal` in several places to handle startup errors. `log.Fatal` calls `os.Exit` internally, which terminates the process immediately without running any `defer`red cleanup — so a `defer cache.Close()` registered earlier in the same function would silently never execute if a later `log.Fatal` fired. This was caught by `golangci-lint`'s `gocritic` linter (`exitAfterDefer`) during a routine commit.

Second, and more significantly: any process supervisor (a container orchestrator stopping a container for a deploy or scale-down, `docker stop`, a process manager) sends `SIGTERM` and expects the process to wind down cleanly within some grace period, not die instantly. A bare `http.ListenAndServe` provides no mechanism to intercept this signal, finish in-flight requests, or stop background goroutines (the reaper, the warm pool maintainers) cleanly — a `SIGTERM` would simply kill the process mid-request, mid-reap, or mid-container-creation.

## Decision

Restructure `main()` into a thin wrapper that calls a `run() error` function, so all setup and the server lifecycle live in a function whose `defer`s execute normally on any return path — `main()` itself only logs and exits on the error `run()` returns, with nothing left to defer at that outer layer.

Within `run()`, use `signal.NotifyContext` to derive a root `context.Context` that is cancelled automatically on `SIGINT`/`SIGTERM`. This context is passed to the reaper and warm pool, whose loops already select on `ctx.Done()`. The HTTP server runs via an explicit `*http.Server` (not the shorthand `ListenAndServe` package function) specifically to gain access to `Shutdown(ctx)`, which stops accepting new connections while allowing in-flight requests up to a bounded timeout to complete, falling back to a hard `Close()` only if graceful shutdown itself times out.

## Consequences

- A `SIGTERM`/`SIGINT` now triggers an orderly sequence: stop accepting new connections, let in-flight requests (including potentially slow ones like `exec` or file transfers) finish within a 15-second window, cancel the shared context so the reaper and warm pool goroutines exit their loops, then close the Redis client via its `defer`, which now reliably executes.
- This is a prerequisite for running Cage under any real container orchestrator (Kubernetes, ECS, Docker Compose with `restart`/`stop` semantics, etc.) — without it, every deploy or scale-down event would risk cutting off in-flight sandbox operations.
- The shutdown timeout (15s) is currently a fixed constant, not configurable via environment variable. If exec operations routinely need longer to complete gracefully, this may need to become tunable — flagged as a small follow-up rather than solved preemptively without evidence it's needed.
- This refactor also fixed several unrelated latent bugs that had accumulated across earlier steps (hardcoded port ignoring `cfg.Port`, hardcoded reaper interval ignoring `cfg.ReaperInterval`, the rate limiter and metrics middleware never actually being registered on the router) — these were caught and corrected as part of the same rewrite rather than addressed separately, since they lived in the same function being restructured.