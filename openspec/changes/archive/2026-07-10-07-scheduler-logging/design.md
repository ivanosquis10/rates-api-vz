# Design: 07-scheduler-logging

## Technical Approach
Automate daily exchange rate scraping using `robfig/cron/v3` set to the `America/Caracas` timezone. The execution hour is configurable via `SCRAPE_HOUR`. Scraping failures are handled by a retry loop with exponential backoff (1m, 2m, 4m). Structured JSON logs are emitted via `log/slog` for each execution attempt, capturing durations and rate details on success, and attempt/error info on failure. Graceful shutdown waits for any running scrape job to complete.

## Architecture Decisions

### Decision: Retry Logic Delay Implementation
- **Choice**: Loop-based execution with `time.After` or context-aware channel selections.
- **Alternatives considered**: Retrying via rescheduled cron jobs or third-party libraries (e.g. `avast/retry-go`).
- **Rationale**: Minimal external dependencies. Channel selection on `ctx.Done()` enables immediate interruption during shutdown.

### Decision: Structured Output Fields
- **Choice**: Standard `log/slog` structured logs outputted to `os.Stdout`.
- **Alternatives considered**: Writing logs to separate file or third-party log exporter.
- **Rationale**: Keeps execution lightweight and adheres to container/cloud native patterns where stdout is aggregated.

## Data Flow
```
[Robfig Cron]
     │ (Triggers at SCRAPE_HOUR)
     ▼
[Scheduler Job] ────────► [executeWithRetry()]
                                 │ (up to 4 attempts)
                                 ▼
                     [RateUsecase.ScrapeRates()]
                                 │
                 ┌───────────────┴───────────────┐
                 ▼                               ▼
      [BCVScraper.Scrape()]            [Repository.SaveRates()]
```
1. Cron triggers execution daily at the Caracas hour specified in config.
2. The scheduler executes the usecase `ScrapeRates` with up to 3 retries (4 total attempts) on failure.
3. If an attempt succeeds, success logs are written per currency rate and retry halts.
4. If an attempt fails, it logs the error/attempt number, waits (1m, 2m, or 4m) and retries unless context is canceled.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Modify | Add `github.com/robfig/cron/v3`. |
| `internal/usecase/rate_usecase.go` | Modify | Update `ScrapeRates` signature to return `([]domain.Rate, error)`. |
| `internal/handler/handler.go` | Modify | Update `Usecaser` interface's `ScrapeRates` signature. |
| `internal/handler/rate_handlers.go` | Modify | Update `TriggerScrape` to use updated `ScrapeRates` signature. |
| `internal/handler/rate_handlers_test.go` | Modify | Update mock usecase implementation and trigger scrape tests. |
| `internal/usecase/rate_usecase_test.go` | Modify | Update unit tests to match `[]domain.Rate` return type. |
| `internal/scheduler/scheduler.go` | New | Implement the background cron scheduler, retry runner, and logging. |
| `internal/scheduler/scheduler_test.go` | New | Integration and scheduler cron setup, retry, and timezone tests. |
| `cmd/api/main.go` | Modify | Wire up, start, and stop the scheduler gracefully. |

## Interfaces / Contracts

### ScrapeRates Method Signature
```go
// internal/usecase/rate_usecase.go
func (uc *RateUsecase) ScrapeRates(ctx context.Context) ([]domain.Rate, error)
```

### Scheduler Type Definition
```go
// internal/scheduler/scheduler.go
type RateScraperUsecase interface {
	ScrapeRates(ctx context.Context) ([]domain.Rate, error)
}

type Scheduler struct {
	cron       *cron.Cron
	usecase    RateScraperUsecase
	scrapeHour int
}
```

### JSON Log Output
- **Success attempt (logged per rate)**:
  `{"time":"...","level":"INFO","msg":"rate scraped successfully","duration_ms":450,"currency":"USD","value":45.2}`
- **Failed attempt**:
  `{"time":"...","level":"ERROR","msg":"scrape attempt failed","error":"...","attempt":1}`

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit / Usecase | Modified Signatures | Assert usecase returns the rates slice on success and wrapped error on failure. |
| Unit / Scheduler | Timezone and Cron | Assert Caracas time location loads correctly and cron expression is correct. |
| Unit / Scheduler | Retry Backoff | Mock usecase failure/success sequence; verify retries run exactly 3 times at 1m, 2m, 4m intervals. |
| Integration | End-to-End Scraper | Trigger a scheduler scrape run, verify SQLite writes, and verify API responses reflect scraped data. |

## Migration / Rollout
No migration required. Database schema remains unchanged.

## Open Questions
- [ ] None
