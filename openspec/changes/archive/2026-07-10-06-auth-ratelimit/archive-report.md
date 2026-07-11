# Archive Report: 06 — Auth & Rate Limiter Middleware

**Change**: 06-auth-ratelimit
**Status**: COMPLETE
**Archived**: 2026-07-10
**Verdict**: PASS (no CRITICAL issues)

---

## Executive Summary

This change secures the Venezuela Rates API by introducing two middleware layers: API Key Authentication and per-IP Rate Limiting. Both middlewares are Chi-compliant and are wired into the main Chi router pipeline. The API Key check is timing-attack resistant, and the Rate Limiter features a concurrent-safe IP map with a context-controlled background janitor to prevent memory and goroutine leaks. The delta specifications have been successfully merged, and all implementation tasks are fully completed and verified.

## What Was Built

| Capability | Impact |
|------------|--------|
| **Timing-Attack Resistant Auth** | Rejects requests with missing or incorrect `X-API-Key` header with HTTP 401 and JSON error. Compares SHA-256 hashes of keys via `subtle.ConstantTimeCompare`. |
| **IP-based Rate Limiter** | Restricts clients based on IP (using `net.SplitHostPort` to strip ports). Employs `golang.org/x/time/rate` token bucket. |
| **HTTP 429 & Retry-After** | Over-limit requests are rejected with HTTP 429 and `Retry-After` header specifying the exact wait time. |
| **Pruning Janitor** | Runs in a background goroutine, scanning and deleting inactive limiters (no activity for 5+ minutes). |
| **Middleware Chaining** | Integrated into the Chi router pipeline in `cmd/api/main.go` in the specified order: Recovery → Logging → RateLimit → Auth. |

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| **SHA-256 Hashing before Compare** | Comparing raw strings of varying lengths leaks the length via timing characteristics. Pre-hashing to 32-byte SHA-256 digests ensures constant-length comparisons. |
| **Atomic Unix lastSeen Timestamp** | Accessing the last activity timestamp concurrently inside request handlers and the janitor can cause a data race. `sync/atomic` operations on an `int64` eliminate this race cleanly. |
| **Context-Aware Janitor** | Accepting the application context in the rate limiter initialization allows stopping the janitor goroutine and ticker upon shutdown, avoiding goroutine leaks during tests. |
| **Split-Phase Map Pruning** | To avoid holding a exclusive write lock over the entire map iteration (which blocks all concurrent requests), the janitor scans under a read lock (`RLock`), flags expired items, and deletes them under a brief write lock (`Lock`). |

## Files Modified

| File | Action | Description |
|------|--------|-------------|
| [internal/middleware/auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) | Created | Timing-resistant API key check middleware |
| [internal/middleware/ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) | Created | Concurrent IP-based rate limiter and cleanup janitor |
| [cmd/api/main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) | Modified | Wired new middlewares after logging in Chi router |
| [internal/middleware/auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go) | Created | Unit tests for valid, missing, and invalid keys |
| [internal/middleware/ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go) | Created | Unit tests for limit capacities, IP separation, janitor pruning, and lifecycle |
| [internal/middleware/integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go) | Created | Test for verifying Chi middleware stack execution order |

## Test Results

- **All passing**: ✅
- **Coverage**:
  - `internal/middleware`: 92.9% statement coverage
- **Spec compliance**: All 6 spec scenarios (including execution order) covered by passing tests

## Engram Observation IDs

| Artifact | Observation ID |
|----------|----------------|
| proposal | #565 |
| spec | #566 |
| design | #567 |
| tasks | #568 |
| verify-report | #570 |

## Lessons Learned

1. **API Key Leaks**: Never compare raw secrets directly using `subtle.ConstantTimeCompare` if they can have different lengths, as the length mismatch can leak key length. Hashing with SHA-256 ensures constant length inputs.
2. **Lock Contention**: Iterating over maps in Go while holding a write lock blocks all concurrent readers. Splitting the logic into a Read Lock scan and a Write Lock delete phase optimizes hot path request latency.
3. **Goroutine Leaks in Tests**: Background tickers/goroutines spawned in libraries or middlewares will leak during test suite execution if not bound to a context that gets cancelled. Always pass a `context.Context` to control their lifecycles.

## Archive Verification

- [x] Delta specifications for `http-server` merged into the main spec file
- [x] Change folder moved to `openspec/changes/archive/2026-07-10-06-auth-ratelimit/`
- [x] Archive contains all artifacts (proposal, design, tasks, verify-report, delta spec)
- [x] All implementation tasks marked complete in `tasks.md`
- [x] Active changes directory no longer contains `06-auth-ratelimit/`
- [x] Verify report shows PASS with no CRITICAL issues

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived.
