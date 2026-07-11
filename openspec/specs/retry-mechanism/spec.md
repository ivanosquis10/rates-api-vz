# Retry Mechanism Specification

## Purpose
Resilience mechanism for scraping failures using retries with exponential backoff.

## Requirements

### Requirement: ExponentialBackoffRetry
On failure of the scrape execution, the system SHALL retry up to 3 additional times. The backoff delays MUST be exactly 1 minute, 2 minutes, and 4 minutes.

#### Scenario: Successful execution on retry
- GIVEN the scraper fails on first attempt
- WHEN the execution is retried
- THEN it waits 1 minute before the second attempt
- AND if the second attempt succeeds, no further retries are performed

#### Scenario: Scraper fails all retries
- GIVEN the scraper fails on all attempts
- WHEN the scheduler runs
- THEN the system retries at 1m, 2m, and 4m intervals
- AND stops after 3 retries (4 total attempts) and logs the final failure
