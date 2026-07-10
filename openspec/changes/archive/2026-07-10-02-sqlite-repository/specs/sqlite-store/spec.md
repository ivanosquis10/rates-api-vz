# Delta for sqlite-store

## MODIFIED Requirements

### Requirement: Repository Implementation

The system SHALL implement the `domain.Repository` interface. `SaveRates` MUST insert all rates in a transaction. `GetLatestRates` MUST return the most recent rate per `(currency, rate_type)` combination. `GetHistoryRates` MUST support filtering by currency, rate_type, and date range with configurable limit.

(Previously: Four scenarios defined but test coverage gaps existed for bank rates, multi-timestamp dedup, rateType filter, date range filter, and public-API duplicate prevention)

#### Scenario: SaveRates persists all rates

- GIVEN 3 Rate values in a slice
- WHEN `repo.SaveRates(ctx, rates)` is called
- THEN 3 rows exist in the `rates` table

#### Scenario: SaveRates persists bank-specific rates

- GIVEN rates with `Bank` set to "Banco de Venezuela" and `Bank` empty (reference)
- WHEN `repo.SaveRates(ctx, rates)` is called
- THEN both rows are persisted with correct `bank` values
- AND reference rates have NULL/empty `bank` in the database

#### Scenario: SaveRates rejects duplicate via public API

- GIVEN a rate with `(currency=USD, rate_type=reference, bank="", scraped_at=T1)` already saved
- WHEN `repo.SaveRates(ctx, [same rate])` is called
- THEN an error is returned
- AND the original row is unchanged

#### Scenario: GetLatestRates returns most recent per type

- GIVEN rates for USD reference at 10:00 and 12:00 on the same day
- WHEN `repo.GetLatestRates(ctx, "USD")` is called
- THEN only the 12:00 rate is returned

#### Scenario: GetLatestRates deduplicates across multiple banks

- GIVEN rates for USD reference from "Banco A" at 10:00 and 12:00, and from "Banco B" at 11:00
- WHEN `repo.GetLatestRates(ctx, "USD")` is called
- THEN only the 12:00 rate from "Banco A" is returned for rate_type=reference
- AND the 11:00 rate from "Banco B" is returned for its combination

#### Scenario: GetHistoryRates respects filters

- GIVEN rates for USD and EUR across multiple days
- WHEN `repo.GetHistoryRates(ctx, "USD", "reference", "2026-07-01", "2026-07-10", 30)` is called
- THEN only USD reference rates within the date range are returned
- AND results are ordered by scraped_at descending
- AND at most 30 results are returned

#### Scenario: GetHistoryRates filters by rateType

- GIVEN USD rates with both "reference" and "parallel" types
- WHEN `repo.GetHistoryRates(ctx, "USD", "reference", "", "", 100)` is called
- THEN only "reference" type rates are returned

#### Scenario: GetHistoryRates filters by date range

- GIVEN USD rates on 2026-07-01, 2026-07-05, and 2026-07-10
- WHEN `repo.GetHistoryRates(ctx, "USD", "", "2026-07-03", "2026-07-08", 100)` is called
- THEN only the 2026-07-05 rate is returned

#### Scenario: GetHistoryRates returns empty slice for no match

- GIVEN no rates match the filters
- WHEN `repo.GetHistoryRates(ctx, ...)` is called
- THEN an empty slice (not nil) is returned with no error
