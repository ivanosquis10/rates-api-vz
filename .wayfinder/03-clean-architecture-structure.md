# Ticket 03: Clean Architecture Folder Structure

**Type**: Grilling (HITL)
**Blocked by**: None (frontier)
**Status**: RESOLVED

## Question

How should the project folders be organized following Clean Architecture?

## Resolution

### Folder tree

```
venezuela-rates-api/
├── cmd/
│   └── api/
│       └── main.go              # Entrypoint, wires DI, starts server
├── internal/
│   ├── domain/
│   │   ├── rate.go              # Rate entity (struct + validation)
│   │   └── errors.go            # Domain error types
│   ├── usecase/
│   │   ├── rate_usecase.go      # Business logic: GetRates, GetHistory, ScrapeRates
│   │   └── rate_usecase_test.go # Unit tests for use cases
│   ├── repository/
│   │   ├── rate_repository.go   # Interface (port)
│   │   └── sqlite/
│   │       └── rate_repo.go     # SQLite implementation (adapter)
│   ├── handler/
│   │   ├── routes.go            # Chi route registration
│   │   ├── rate_handler.go      # HTTP handlers: GET /rates, GET /rates/history
│   │   └── response.go          # JSON response helpers, error envelope
│   ├── middleware/
│   │   ├── auth.go              # API Key validation
│   │   └── ratelimit.go         # Rate limiter middleware
│   ├── scraper/
│   │   └── bcv_scraper.go       # BCV scraping logic (goquery)
│   ├── scheduler/
│   │   └── scheduler.go         # Daily cron/ticker for scraping
│   └── config/
│       └── config.go            # Env var loading + validation
├── go.mod
├── go.sum
├── .env.example                 # Documented env vars
└── README.md
```

### Layer responsibilities

```
                  ┌─────────────────────────────┐
                  │       handler (HTTP)         │  ← Chi routes, request parsing
                  │       middleware (auth, RL)   │  ← Cross-cutting concerns
                  └──────────────┬──────────────┘
                                 │ calls
                  ┌──────────────▼──────────────┐
                  │       usecase (business)     │  ← Orchestration, validation
                  │       rate_usecase.go        │  ← No HTTP/DB details
                  └──────────────┬──────────────┘
                                 │ uses interface
              ┌──────────────────┼──────────────────┐
              │                  │                  │
    ┌─────────▼─────────┐ ┌─────▼─────┐ ┌─────────▼─────────┐
    │ repository (port) │ │  scraper  │ │    scheduler      │
    │ interface         │ │  (BCV)    │ │    (cron)         │
    └─────────┬─────────┘ └───────────┘ └───────────────────┘
              │ implements
    ┌─────────▼─────────┐
    │   sqlite (adapter)│  ← Actual SQLite queries
    └───────────────────┘
```

### Key design decisions

1. **Interfaces in `domain`** — `RateRepository` interface lives in `domain/`, not `repository/`. This keeps the domain pure and lets use cases depend on abstractions.

2. **Implementations in sub-packages** — `repository/sqlite/` implements the interface. Swapping to Postgres means adding `repository/postgres/`, not changing use cases.

3. **Handler receives usecase interface** — Handlers don't know about SQLite. They receive `RateUseCase` interface, making them testable with mocks.

4. **Manual DI in `main.go`** — Wire everything in main: config → sqlite repo → usecase → handler → router → server. Clear, explicit, no magic.

5. **Env vars only** — `config.go` loads from `os.Getenv`, validates required vars, fails fast on startup. No .env files in production.

### Dependency flow (rules)

```
domain      → depends on NOTHING (pure structs + interfaces)
usecase     → depends on domain (interfaces from domain)
repository  → depends on domain (implements domain interfaces)
handler     → depends on domain + usecase (calls usecase, returns domain types)
scraper     → depends on domain (returns domain types)
scheduler   → depends on usecase (calls ScrapeRates)
middleware  → depends on nothing (pure functions)
config      → depends on nothing (reads env)
```

### What goes where

| Concern | Package | Why |
|---------|---------|-----|
| Rate struct | `domain/rate.go` | Core entity, used everywhere |
| Business rules | `usecase/rate_usecase.go` | Orchestration, validation |
| SQL queries | `repository/sqlite/rate_repo.go` | Implementation detail |
| HTTP routes | `handler/routes.go` | Chi setup |
| Request/Response | `handler/rate_handler.go` + `response.go` | HTTP layer |
| API Key check | `middleware/auth.go` | Cross-cutting |
| Rate limiting | `middleware/ratelimit.go` | Cross-cutting |
| BCV scraping | `scraper/bcv_scraper.go` | External integration |
| Daily job | `scheduler/scheduler.go` | Time-based trigger |
| Env vars | `config/config.go` | Configuration |
| DI wiring | `cmd/api/main.go` | Composition root |
