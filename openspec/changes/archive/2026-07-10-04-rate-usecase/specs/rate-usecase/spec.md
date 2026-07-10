# Rate Usecase Specification

## Purpose

Business logic orchestration layer that coordinates BCV scraping and SQLite persistence. Provides a testable interface for the API layer to interact with domain operations without knowing implementation details.

## Requirements

### Requirement: RateUsecase Struct

The system SHALL define a `RateUsecase` struct with two dependencies: `domain.Repository` and `scraper.Scraper`.

A constructor `NewRateUsecase(repo, scraper)` MUST accept both dependencies and return a pointer to `RateUsecase`.

#### Scenario: Constructor with valid dependencies

- GIVEN a `domain.Repository` and a `scraper.Scraper`
- WHEN `NewRateUsecase` is called
- THEN a non-nil `RateUsecase` is returned
- AND both dependencies are stored for use by subsequent methods

### Requirement: ScrapeRates

The `ScrapeRates(ctx)` method SHALL call `scraper.Scraper.Scrape(ctx)`, then persist the resulting rates via `domain.Repository.SaveRates(ctx, rates)`.

On success, it SHALL return the count of persisted rates and a nil error.

On scraper failure, it SHALL log the error with `slog` and return a wrapped error. On repository failure after a successful scrape, it SHALL log the error with `slog` and return a wrapped error. The scraper and repository errors MUST be distinguishable by the caller via `%w` wrapping.

#### Scenario: Successful scrape and persist

- GIVEN a scraper that returns 5 rates
- WHEN `ScrapeRates` is called
- THEN `SaveRates` is called with the 5 rates
- AND the method returns (5, nil)

#### Scenario: Scraper returns error

- GIVEN a scraper that returns an error
- WHEN `ScrapeRates` is called
- THEN `SaveRates` is NOT called
- AND the method returns (0, wrapped error)
- AND the error is logged via slog

#### Scenario: Repository save fails after successful scrape

- GIVEN a scraper that returns 3 rates and a repository that fails on `SaveRates`
- WHEN `ScrapeRates` is called
- THEN `SaveRates` is called with the 3 rates
- AND the method returns (0, wrapped error)
- AND the error is logged via slog

### Requirement: GetCurrentRates

The `GetCurrentRates(ctx, currency, rateType)` method SHALL call `domain.Repository.GetLatestRates(ctx, currency)` and return the resulting slice.

If `rateType` is non-empty, the returned slice MUST be filtered to include only rates whose `RateType` matches. Empty `rateType` returns all types.

#### Scenario: Get latest rates for a currency

- GIVEN the repository has rates for USD
- WHEN `GetCurrentRates(ctx, "USD", "")` is called
- THEN the method returns all latest USD rates (reference, buy, sell)

#### Scenario: Filter by rate type

- GIVEN the repository has reference, buy, and sell rates for USD
- WHEN `GetCurrentRates(ctx, "USD", "reference")` is called
- THEN the returned slice contains only rates where `RateType == "reference"`

#### Scenario: No rates exist

- GIVEN the repository returns an empty slice for a currency
- WHEN `GetCurrentRates` is called
- THEN the method returns an empty slice (not nil) and a nil error

### Requirement: GetHistoryRates

The `GetHistoryRates(ctx, currency, rateType, from, to, limit)` method SHALL delegate to `domain.Repository.GetHistoryRates(ctx, currency, rateType, from, to, limit)` and return the result directly.

#### Scenario: Retrieve history with all filters

- GIVEN the repository has historical USD rates
- WHEN `GetHistoryRates(ctx, "USD", "buy", "2026-01-01", "2026-07-01", 50)` is called
- THEN the method returns up to 50 matching rates ordered by `ScrapedAt` DESC

#### Scenario: Empty history

- GIVEN the repository returns an empty slice for the given filters
- WHEN `GetHistoryRates` is called
- THEN the method returns an empty slice (not nil) and a nil error

### Requirement: Logging

All errors occurring within `RateUsecase` methods MUST be logged with `slog.Error` before being returned. Log messages MUST include the method name and a contextual description.

#### Scenario: Error logging includes context

- GIVEN a scraper error occurs during `ScrapeRates`
- WHEN the error is logged
- THEN the log entry includes `"method", "ScrapeRates"` and the error description
