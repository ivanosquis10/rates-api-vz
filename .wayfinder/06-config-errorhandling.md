# Ticket 06: Configuration & Error Handling

**Type**: Grilling (HITL)
**Blocked by**: None (frontier)
**Status**: RESOLVED

## Question

How should configuration be managed and errors handled across the API?

## Resolution

### Configuration (.env + env vars)

```go
// config/config.go
package config

type Config struct {
    Port        int    `env:"PORT" envDefault:"8080"`
    DBPath      string `env:"DB_PATH" envDefault:"./rates.db"`
    APIKey      string `env:"API_KEY" envRequired:"true"`
    ScrapeHour  int    `env:"SCRAPE_HOUR" envDefault:"8"`
    RateLimit   int    `env:"RATE_LIMIT" envDefault:"60"`
}

func Load() (*Config, error) {
    // Load from .env file if present (local dev)
    // Environment variables override .env (production)
    // Fail fast if required vars missing
}
```

**Configuration approach:**
- `.env` file for local development (gitignored)
- `.env.example` documents all variables with descriptions and defaults
- Environment variables override `.env` values (production)
- Library: `joho/godotenv` for .env loading + `os.Getenv` for reading

**Required env vars:**
| Var | Required | Default | Description |
|-----|----------|---------|-------------|
| `PORT` | No | 8080 | Server port |
| `DB_PATH` | No | `./rates.db` | SQLite database path |
| `API_KEY` | **Yes** | — | API key for authentication |
| `SCRAPE_HOUR` | No | 8 | Hour to run daily scraping (0-23) |
| `RATE_LIMIT` | No | 60 | Requests per minute per IP |

### Error Response Envelope

```go
// handler/response.go
package handler

type ErrorResponse struct {
    Error ErrorBody `json:"error"`
}

type ErrorBody struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func respondWithError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error: ErrorBody{
            Code:    httpStatusToCode(status),
            Message: message,
        },
    })
}

func httpStatusToCode(status int) string {
    switch status {
    case 400: return "BAD_REQUEST"
    case 401: return "UNAUTHORIZED"
    case 404: return "NOT_FOUND"
    case 429: return "RATE_LIMITED"
    case 500: return "INTERNAL_ERROR"
    default:  return "ERROR"
    }
}
```

**Examples:**
```json
// 401 Unauthorized
{ "error": { "code": "UNAUTHORIZED", "message": "Invalid or missing API key" } }

// 404 Not Found
{ "error": { "code": "NOT_FOUND", "message": "No rates found for the specified date" } }

// 429 Rate Limited
{ "error": { "code": "RATE_LIMITED", "message": "Rate limit exceeded. Try again in 60 seconds" } }

// 500 Internal Error
{ "error": { "code": "INTERNAL_ERROR", "message": "An unexpected error occurred" } }
```

### Logging (slog)

```go
// cmd/api/main.go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)
    
    // Usage:
    slog.Info("scrape completed",
        "currency", "USD",
        "rate", 709.69,
        "duration_ms", 1200,
    )
    
    slog.Error("scrape failed",
        "error", err,
        "attempt", 3,
        "url", "https://www.bcv.org.ve",
    )
}
```

**Log format:** JSON structured logs to stdout (container-friendly).

### Error handling patterns

| Layer | Error handling |
|-------|---------------|
| **Handler** | Catch usecase errors, map to HTTP status codes |
| **Usecase** | Return domain errors, don't leak internals |
| **Repository** | Wrap SQL errors, return domain errors |
| **Scraper** | Return timeout/network errors with context |

```go
// Usecase returns domain errors
var (
    ErrRateNotFound = errors.New("rate not found")
    ErrScrapeFailed = errors.New("scrape failed")
)

// Handler maps domain errors to HTTP
func handler(uc RateUseCase) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rates, err := uc.GetCurrentRates(r.Context())
        if err != nil {
            switch {
            case errors.Is(err, usecase.ErrRateNotFound):
                respondWithError(w, 404, "No rates available")
            default:
                slog.Error("get rates failed", "error", err)
                respondWithError(w, 500, "Internal server error")
            }
            return
        }
        respondWithJSON(w, 200, rates)
    }
}
```
