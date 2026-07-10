# Proposal: Rate Usecase (Business Logic)

## Intent

Add a business logic orchestration layer that coordinates scraping and persistence operations while maintaining clean separation of concerns. The usecase layer provides a testable interface for the API layer to interact with domain operations without knowing implementation details.

## Scope

### In Scope
- `RateUsecase` struct implementing business logic for rate operations
- `ScrapeRates(ctx)` method: calls Scraper interface, persists results via Repository interface
- `GetCurrentRates(ctx, currency, rateType)` method: retrieves latest rates with optional filters
- `GetHistoryRates(ctx, currency, rateType, from, to, limit)` method: retrieves historical rates with all filters
- Graceful error handling with slog logging for scraper and repository failures
- Manual mock implementations (mockScraper, mockRepository) for unit testing
- Unit tests covering: scrape success, scrape failure, get current with filters, get history with filters, empty results

### Out of Scope
- HTTP handlers or API endpoints (handled in separate change)
- Rate validation or business rules beyond delegation to interfaces
- Caching or rate limiting
- Pagination beyond limit parameter
- Authentication or authorization

## Capabilities

### New Capabilities
- `rate-usecase`: Business logic orchestration layer for rate scraping and retrieval operations

### Modified Capabilities
- None

## Approach

Create `internal/usecase/rate_usecase.go` with a `RateUsecase` struct that depends on `domain.Repository` and `scraper.Scraper` interfaces. The struct will:
- Accept both dependencies via constructor `NewRateUsecase(repo Repository, scraper Scraper)`
- Implement three methods matching the acceptance criteria
- Use `slog` for structured logging of errors
- Return domain errors wrapped with context for API layer handling

Tests in `internal/usecase/rate_usecase_test.go` will use manual mock implementations that satisfy the interfaces without external dependencies.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/usecase/rate_usecase.go` | New | Business logic layer implementation |
| `internal/usecase/rate_usecase_test.go` | New | Unit tests with manual mocks |
| `internal/usecase/doc.go` | New | Package documentation |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Interface drift between usecase and domain | Low | Follow existing interface signatures exactly |
| Test coverage gaps | Medium | Explicit test cases per acceptance criteria |
| Error wrapping verbosity | Low | Use consistent `%w` wrapping pattern |

## Rollback Plan

Delete the `internal/usecase/` directory and remove any imports referencing it. The change is purely additive with no modifications to existing code.

## Dependencies

- `internal/domain/repository.go` — Repository interface (completed in #3)
- `internal/scraper/scraper.go` — Scraper interface (completed in #4)
- `internal/domain/rate.go` — Rate struct (completed in #3)

## Success Criteria

- [ ] `RateUsecase` struct compiles with correct interface dependencies
- [ ] `ScrapeRates` calls scraper and persists results atomically
- [ ] `GetCurrentRates` returns latest rates filtered by currency and rateType
- [ ] `GetHistoryRates` returns historical rates with all filter parameters
- [ ] Scraper errors are logged and returned with context
- [ ] Repository errors are logged and returned with context
- [ ] Unit tests pass with manual mocks (no external dependencies)
- [ ] Test coverage includes: success, failure, filters, empty results
- [ ] `go test ./internal/usecase/...` succeeds
