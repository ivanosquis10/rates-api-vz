<Proposal: 07-scheduler-logging>
## Intent
Enable automated daily scraping of exchange rates at a configurable hour with robust error retry logic and structured logging to stdout.

## Scope
### In Scope
- Implement a background cron scheduler using `robfig/cron/v3` with the `America/Caracas` timezone.
- Load `SCRAPE_HOUR` env var (default: 8) and schedule the cron.
- Implement exponential backoff retry logic (3 attempts: 1m, 2m, 4m) for scraping failures.
- Structured JSON logging to stdout using standard `log/slog` for success and failure attempts.
- Integration test covering the flow from scraping to database storage to API handler response.
- Graceful startup and shutdown of the scheduler aligned with the HTTP server.

### Out of Scope
- Modifying the web scraping selector logic itself.
- Setting up persistent external log storage (e.g., Elasticsearch, Loki).
- Adding notification services (e.g., Email, Slack) for scraper failures.

## Capabilities
### New Capabilities
- `Scraper Scheduler`: Automatically runs daily scrapers at a specified Caracas time.
- `Retry Mechanism`: Automatically retries failed scraping tasks up to 3 times with exponential backoff.
- `Structured JSON Scraper Logs`: Emits standard structured logs for visibility into cron executions.

### Modified Capabilities
- `ScrapeRates Usecase`: Modified to return the list of scraped rates (`[]domain.Rate`) for logging.
- `API Server Entrypoint`: Initialize and gracefully terminate the scheduler alongside HTTP server.

## Approach
1. Add `robfig/cron/v3` to `go.mod`.
2. Update `RateUsecase.ScrapeRates` signature to return `([]domain.Rate, error)`.
3. Implement `internal/scheduler/scheduler.go` with a `Scheduler` struct that initializes the cron, manages timezones, executes with retry backoff, and logs success/failure.
4. Integrate the scheduler in `cmd/api/main.go` and handle graceful stop via OS signals.
5. Add integration tests in `internal/scheduler/scheduler_test.go` to mock the scraping/database/API cycle.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/config.go` | Modified | Ensure `SCRAPE_HOUR` is loaded (already present). |
| `internal/usecase/rate_usecase.go` | Modified | Change `ScrapeRates` return type to `([]domain.Rate, error)`. |
| `internal/handler/handler.go` | Modified | Update interface method signature for `ScrapeRates`. |
| `internal/handler/rate_handlers.go` | Modified | Update handlers to match new usecase signature. |
| `internal/scheduler/scheduler.go` | New | Main scheduler logic, cron setups, logging and retries. |
| `internal/scheduler/scheduler_test.go` | New | Integration and unit tests for the scheduler. |
| `cmd/api/main.go` | Modified | Start and stop the scheduler on application lifecycle. |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Cron overlaps if task runs too long | Low | The scrapers run quickly; cron runs daily. We will track execution status. |
| Scraper block/failure | Medium | Handled by 3 retries with 1m, 2m, 4m exponential backoff. |

## Rollback Plan
Revert changes via Git, remove the scheduler invocation in `main.go`, and restore original `ScrapeRates` signatures.

## Dependencies
- `github.com/robfig/cron/v3`

## Success Criteria
- [ ] Scheduler triggers scraper successfully at the configured `SCRAPE_HOUR` in Caracas timezone.
- [ ] Failed scrapes trigger exactly 3 retries with 1m, 2m, 4m backoff.
- [ ] Structured JSON logs are emitted on success (with currency, value, duration) and failure.
- [ ] Integration tests pass successfully.
</Proposal: 07-scheduler-logging>
