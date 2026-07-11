# Scraper Scheduler Specification

## Purpose
Automated background scheduler for executing daily scrapers at a configurable hour in Caracas timezone.

## Requirements

### Requirement: ChronScheduling
The system SHALL use `robfig/cron/v3` with `America/Caracas` timezone. It SHALL read `SCRAPE_HOUR` env var (defaulting to 8) and schedule daily execution at that hour.

#### Scenario: Normal Scheduler Startup
- GIVEN SCRAPE_HOUR is set to 9
- WHEN the scheduler starts
- THEN a cron job is scheduled for "0 9 * * *" in America/Caracas timezone

### Requirement: GracefulShutdown
The scheduler SHALL support graceful shutdown. When stopped, it SHALL wait for any currently running scrape job to finish before exiting.

#### Scenario: Stop Scheduler
- GIVEN the scheduler is running a job
- WHEN the stop signal is received
- THEN the scheduler blocks until the running job completes and then terminates
