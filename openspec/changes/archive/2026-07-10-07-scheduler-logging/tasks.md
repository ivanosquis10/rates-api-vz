# Tasks: 07-scheduler-logging

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Medium

| Field | Value |
|-------|-------|
| Estimated changed lines | ~350 lines |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

### Suggested Work Units
| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| U1 | Signature Updates & Mock Tests | PR 1 | Update usecase, interfaces, mock, handlers |
| U2 | Scheduler, Logging, Wiring & Main | PR 1 | Implement scheduler, cron setup, retry, wire in main.go |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 Add `github.com/robfig/cron/v3` to [go.mod](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/go.mod) and run go mod tidy.
- [x] 1.2 Import `_ "time/tzdata"` in [cmd/api/main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) or [scheduler.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scheduler/scheduler.go) to ensure Windows timezone DB support.

## Phase 2: Core Implementation
- [x] 2.1 Update [ScrapeRates](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/usecase/rate_usecase.go) signature to return `([]domain.Rate, error)`.
- [x] 2.2 Implement background [Scheduler](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scheduler/scheduler.go) and structured log output.
- [x] 2.3 Use `time.NewTimer` and `timer.Stop()` in the retry backoff loop instead of `time.After` to prevent timer resource leaks.
- [x] 2.4 Propagate the cancellation context to the retry loop to cancel immediately during shutdown without delay.

## Phase 3: Integration / Wiring
- [x] 3.1 Update [Usecaser](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/handler.go) interface to match the updated [ScrapeRates](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/usecase/rate_usecase.go) signature.
- [x] 3.2 Update [TriggerScrape](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers.go) to keep count response compatible using `len(rates)`.
- [x] 3.3 Wire up the scheduler in [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go). Block gracefully on cron stop using `<-scheduler.Stop().Done()`.

## Phase 4: Testing / Verification
- [x] 4.1 Update mock usecase implementation and trigger scrape tests in [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go).
- [x] 4.2 Update unit tests in [rate_usecase_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/usecase/rate_usecase_test.go).
- [x] 4.3 Create scheduler unit/timezone and backoff tests in [scheduler_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scheduler/scheduler_test.go).

## Phase 5: Cleanup
- [x] 5.1 Run `go test ./...` to verify all tests pass and code compiles cleanly.
