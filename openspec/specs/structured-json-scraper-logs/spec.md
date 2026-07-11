# Structured JSON Scraper Logs Specification

## Purpose
Provides structured JSON logging using `log/slog` for scheduler executions to allow tracking of success, duration, and failures.

## Requirements

### Requirement: StructuredLogging
The system SHALL log scraper execution attempts to stdout in JSON format using `log/slog`.
On success, the log MUST contain the keys: `msg`, `duration_ms` (execution time), and details of the rates (including `currency` and `value`).
On failure, the log MUST contain: `msg`, `error`, and `attempt` number.

#### Scenario: Successful scrape log
- GIVEN a successful scraping job returning USD rate 45.2
- WHEN the log is written
- THEN the JSON output contains "msg", "duration_ms", "currency": "USD", and "value": 45.2

#### Scenario: Failed scrape log
- GIVEN a failed scraping job at attempt 2
- WHEN the error is logged
- THEN the JSON output contains "msg", "error", and "attempt": 2
