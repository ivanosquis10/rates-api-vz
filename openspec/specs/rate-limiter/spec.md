# Rate Limiter Specification

## Purpose

Define the structure, initialization, and background lifecycle cleanup of the RateLimiter middleware component.

## Requirements

### Requirement: RateLimiter Struct and Construction

The package `internal/middleware` SHALL expose a public `RateLimiter` struct and a constructor `NewRateLimiter(ctx context.Context, limitPerMin int) *RateLimiter`. The constructor MUST initialize the underlying client storage map and start a background janitor routine using the provided context.

#### Scenario: Constructor returns initialized RateLimiter
- GIVEN a valid background context and a limit of 100 requests per minute
- WHEN NewRateLimiter is called
- THEN a non-nil *RateLimiter instance is returned
- AND the background janitor begins running

### Requirement: Janitor Background Lifecycle

The rate limiter's background janitor goroutine SHALL monitor inactive clients. The janitor MUST run periodically, and it MUST prune client limiters that have been inactive for 5 minutes or more. The janitor MUST terminate gracefully when the context passed during construction is cancelled.

#### Scenario: Inactive clients are pruned
- GIVEN a client limiter with last activity older than 5 minutes
- WHEN the background janitor runs
- THEN the inactive client is removed from the active clients map

#### Scenario: Janitor terminates on context cancellation
- GIVEN a running rate limiter janitor
- WHEN the parent context is cancelled
- THEN the janitor goroutine exits cleanly
