# Rate Usecase Delta Specification

## Modified Requirements

### Requirement: ScrapeRates

The `ScrapeRates(ctx)` method SHALL call `scraper.Scraper.Scrape(ctx)`, then persist the resulting rates via `domain.Repository.SaveRates(ctx, rates)`.

On success, it SHALL return the slice of rates and a nil error.
(Previously: On success, it SHALL return the count of persisted rates and a nil error.)

On scraper failure, it SHALL log the error with `slog` and return a wrapped error. On repository failure after a successful scrape, it SHALL log the error with `slog` and return a wrapped error. The scraper and repository errors MUST be distinguishable by the caller via `%w` wrapping.

#### Scenario: Successful scrape and persist

- GIVEN a scraper that returns 5 rates
- WHEN `ScrapeRates` is called
- THEN `SaveRates` is called with the 5 rates
- AND the method returns (rates, nil)
(Previously: - AND the method returns (5, nil))

#### Scenario: Scraper returns error

- GIVEN a scraper that returns an error
- WHEN `ScrapeRates` is called
- THEN `SaveRates` is NOT called
- AND the method returns (nil, wrapped error)
(Previously: - AND the method returns (0, wrapped error))
- AND the error is logged via slog

#### Scenario: Repository save fails after successful scrape

- GIVEN a scraper that returns 3 rates and a repository that fails on `SaveRates`
- WHEN `ScrapeRates` is called
- THEN `SaveRates` is called with the 3 rates
- AND the method returns (nil, wrapped error)
(Previously: - AND the method returns (0, wrapped error))
- AND the error is logged via slog
