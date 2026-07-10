<div align="center">

# 🇻🇪 Venezuela Rates API

**Exchange rates from Banco Central de Venezuela, automated and accessible.**

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Issues](https://img.shields.io/github/issues/ivanosquis10/api-rates-venezuela?style=flat-square)](https://github.com/ivanosquis10/api-rates-venezuela/issues)

</div>

---

## What is this?

A Go API that **scrapes daily exchange rates** (USD & EUR) from the [Banco Central de Venezuela](https://www.bcv.org.ve) website, stores them in **SQLite**, and exposes them through authenticated HTTP endpoints.

Built with **Clean Architecture**, **Chi router**, and **unit tests**.

## Features

- 🔄 **Automatic daily scraping** — configured hour, Caracas timezone
- 📊 **Two rate sources** — BCV reference (weighted average) + bank buy/sell rates
- 🗄️ **SQLite storage** — lightweight, zero-config database
- 🔐 **API Key authentication** — constant-time comparison, internal use
- ⚡ **Rate limiting** — per-IP token bucket, configurable
- 📈 **Historical data** — query rates by date range, currency, type
- 🧪 **Tested** — unit tests, repository tests with SQLite in-memory
- 📝 **Structured logging** — JSON logs via `slog`

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25 |
| HTTP Router | [Chi](https://github.com/go-chi/chi) |
| Database | SQLite ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)) |
| Scraping | [goquery](https://github.com/PuerkitoBio/goquery) |
| Scheduler | [robfig/cron](https://github.com/robfig/cron) |
| Rate Limiter | [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) |
| Logging | `log/slog` (stdlib) |

## Architecture

```
cmd/api/main.go          → Entrypoint, dependency injection
internal/
  domain/                 → Entities, interfaces (pure, no deps)
  usecase/                → Business logic
  repository/sqlite/      → SQLite implementation
  handler/                → HTTP handlers + routes
  middleware/              → Auth, rate limiter
  scraper/                → BCV scraping logic
  scheduler/              → Daily cron job
  config/                 → Environment variable loading
```

**Dependency rule**: `domain` depends on nothing. Everything depends on `domain`.

## Getting Started

### Prerequisites

- Go 1.25+
- GCC (for CGO, if using non-pure-Go SQLite driver)

### Installation

```bash
git clone https://github.com/ivanosquis10/api-rates-venezuela.git
cd api-rates-venezuela
go mod download
```

### Configuration

All configuration is via environment variables. Copy the example file and fill in your values:

```bash
cp .env.example .env
```

Required variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | Server port |
| `DB_PATH` | No | `./rates.db` | SQLite database path |
| `API_KEY` | **Yes** | — | API key for authentication |
| `SCRAPE_HOUR` | No | `8` | Hour to run daily scraping (0-23) |
| `RATE_LIMIT` | No | `60` | Requests per minute per IP |

### Run

```bash
# Set required env var
export API_KEY="your-secret-key"

# Run the server
go run cmd/api/main.go
```

## API Endpoints

All endpoints require the `X-API-Key` header.

### Get Current Rates

```http
GET /rates?currency=USD&type=reference
```

| Param | Type | Description |
|-------|------|-------------|
| `currency` | string | Filter by `USD` or `EUR` |
| `type` | string | Filter by `reference`, `buy`, or `sell` |

**Response:**
```json
{
  "data": [
    {
      "currency": "USD",
      "rate_type": "reference",
      "bank": null,
      "value": 709.69,
      "scraped_at": "2026-07-10T08:00:00-04:00"
    }
  ]
}
```

### Get Historical Rates

```http
GET /rates/history?currency=USD&from=2026-07-01&to=2026-07-10&limit=30
```

| Param | Type | Description |
|-------|------|-------------|
| `currency` | string | Filter by currency |
| `type` | string | Filter by rate type |
| `from` | string | Start date (YYYY-MM-DD) |
| `to` | string | End date (YYYY-MM-DD) |
| `limit` | int | Max results (default: 100) |

### Trigger Scrape (Admin)

```http
POST /admin/scrape
```

Returns immediately. Scraping runs in the background.

### Error Response

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing API key"
  }
}
```

| Status | Code | When |
|--------|------|------|
| 401 | `UNAUTHORIZED` | Missing or invalid API key |
| 429 | `RATE_LIMITED` | Too many requests ( Retry-After header included) |
| 404 | `NOT_FOUND` | No rates found |
| 500 | `INTERNAL_ERROR` | Unexpected server error |

## Project Status

This project is in **planning phase**. All architecture decisions are documented in `.wayfinder/`.

### Implementation Tickets

| # | Ticket | Status |
|---|--------|--------|
| 1 | [Project Scaffolding](https://github.com/ivanosquis10/api-rates-venezuela/issues/2) | 🔲 Ready |
| 2 | [SQLite Repository](https://github.com/ivanosquis10/api-rates-venezuela/issues/3) | 🔲 Blocked |
| 3 | [BCV Scraper](https://github.com/ivanosquis10/api-rates-venezuela/issues/4) | 🔲 Blocked |
| 4 | [Rate Usecase](https://github.com/ivanosquis10/api-rates-venezuela/issues/5) | 🔲 Blocked |
| 5 | [HTTP Server & Endpoints](https://github.com/ivanosquis10/api-rates-venezuela/issues/6) | 🔲 Blocked |
| 6 | [Auth & Rate Limiter](https://github.com/ivanosquis10/api-rates-venezuela/issues/7) | 🔲 Blocked |
| 7 | [Scheduler & Logging](https://github.com/ivanosquis10/api-rates-venezuela/issues/8) | 🔲 Blocked |

## License

MIT
