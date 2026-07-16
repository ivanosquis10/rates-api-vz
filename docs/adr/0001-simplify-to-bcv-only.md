# ADR 0001: Simplify Rate Scraper and Schema to BCV Official Rates Only

## Status
Accepted

## Context
The current application scrapes and stores official reference rates (USD and EUR) alongside commercial bank-specific buy/sell rates (e.g., BBVA Provincial, Banesco, etc.) from the Banco Central de Venezuela (BCV). This design introduces excessive complexity, high database growth, and redundant data that is not required for the target MVP.

We want to align the application design and API contract with the minimalist and standard model used by `dolarapi.com`.

## Decision
1. **Reduce Scraper Scope**: Drop the commercial bank table parsing completely. The scraper will only fetch the official reference rates for USD and EUR from the BCV homepage.
2. **Database Schema Simplification**: Modify the `rates` database schema to store only `currency`, `value`, and `scraped_at`. Remove the `bank` and `rate_type` columns (since all rates are official references).
3. **API Redesign**: Introduce RESTful endpoints in English matching the `dolarapi.com` structure:
   - `GET /v1/dollars` -> Returns an array containing the latest official USD rate.
   - `GET /v1/dollars/official` -> Returns the single latest official USD rate.
   - `GET /v1/euros` -> Returns an array containing the latest official EUR rate.
   - `GET /v1/euros/official` -> Returns the single latest official EUR rate.
   - `GET /v1/history/dollars` -> Returns historical official USD rates.
   - `GET /v1/history/euros` -> Returns historical official EUR rates.
4. **Scraping Strategy (Cron)**: Replace the single daily scrape schedule with a periodic checking strategy that keeps the rates updated without overloading the BCV server.

## Consequences
- Scraper logic will be simpler and less prone to breakage from HTML changes in the bank table.
- Database size and query complexity will be drastically reduced.
- API consumers will receive standard, clean JSON payloads.
