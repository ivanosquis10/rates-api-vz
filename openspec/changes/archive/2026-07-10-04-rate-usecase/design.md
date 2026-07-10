# Design: 04 — Rate Usecase (Business Logic)

## Technical Approach

Add a thin orchestration layer in `internal/usecase/` that coordinates `scraper.Scraper` and `domain.Repository` behind a clean struct. The usecase owns no business rules beyond delegation and logging — it's a pass-through that makes the API layer testable without real I/O. Follows the same dependency-injection-via-constructor pattern used by `store.NewStore` and `scraper.NewBCVScraper`.

## Architecture Decisions

### Decision: Manual mocks over interfaces in tests

**Choice**: Hand-rolled `mockScraper` and `mockRepository` structs that implement the interfaces.
**Alternatives considered**: testify/mock, gomock, or a mock generation tool.
**Rationale**: The project has zero test dependencies beyond the stdlib. Adding a mock library for 2 small interfaces violates YAGNI. Manual mocks are explicit, debuggable, and match the project's minimalist style.

### Decision: In-package filter for rateType in GetCurrentRates

**Choice**: After calling `repo.GetLatestRates`, filter the returned slice in Go by `rateType` when non-empty.
**Alternatives considered**: Push filtering into the Repository interface.
**Rationale**: The Repository's `GetLatestRates` signature is already defined in change #2. Adding a `rateType` param would require modifying a completed interface. Filtering in-memory on a small slice (typically <10 items) is trivial and keeps the interface boundary clean.

### Decision: No domain error wrapping beyond %w

**Choice**: Log with `slog.Error` then return `fmt.Errorf("method: %w", err)`.
**Alternatives considered**: Custom domain error types per method.
**Rationale**: The spec requires distinguishable errors via `%w`. The existing `domain/errors.go` sentinel errors are already sufficient. The API layer (future change) will handle logging and mapping to HTTP codes. The usecase only needs to preserve the error chain.

## Data Flow

```
API Layer (future) ──→ RateUsecase
                            │
            ┌───────────────┴───────────────┐
            ▼                               ▼
    scraper.Scraper                  domain.Repository
      .Scrape(ctx)                  .GetLatestRates(ctx, currency)
            │                       .GetHistoryRates(ctx, ...)
            ▼                               ▼
      []domain.Rate ◄─────────── (returned or passed through)
```

**ScrapeRates flow:**
```
ScrapeRates(ctx)
    → scraper.Scrape(ctx)
        → on error: slog.Error + return (0, err)
    → repo.SaveRates(ctx, rates)
        → on error: slog.Error + return (0, err)
    → return (len(rates), nil)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/usecase/doc.go` | Create | Package documentation |
| `internal/usecase/rate_usecase.go` | Create | `RateUsecase` struct + 3 methods |
| `internal/usecase/rate_usecase_test.go` | Create | Unit tests with manual mocks |

## Interfaces / Contracts

```go
// internal/usecase/rate_usecase.go
package usecase

import (
    "context"
    "fmt"
    "log/slog"
    "strings"

    "github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

// RateUsecase orchestrates rate scraping and retrieval.
type RateUsecase struct {
    repo    domain.Repository
    scraper scraper.Scraper
}

// NewRateUsecase creates a RateUsecase with the given dependencies.
func NewRateUsecase(repo domain.Repository, s scraper.Scraper) *RateUsecase {
    return &RateUsecase{repo: repo, scraper: s}
}

// ScrapeRates calls the scraper and persists results.
func (uc *RateUsecase) ScrapeRates(ctx context.Context) (int, error) {
    rates, err := uc.scraper.Scrape(ctx)
    if err != nil {
        slog.Error("scrape failed", "method", "ScrapeRates", "error", err)
        return 0, fmt.Errorf("ScrapeRates scrape: %w", err)
    }

    if err := uc.repo.SaveRates(ctx, rates); err != nil {
        slog.Error("save rates failed", "method", "ScrapeRates", "error", err)
        return 0, fmt.Errorf("ScrapeRates save: %w", err)
    }

    return len(rates), nil
}

// GetCurrentRates returns latest rates, optionally filtered by rateType.
func (uc *RateUsecase) GetCurrentRates(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
    rates, err := uc.repo.GetLatestRates(ctx, currency)
    if err != nil {
        slog.Error("get latest rates failed", "method", "GetCurrentRates", "error", err)
        return nil, fmt.Errorf("GetCurrentRates: %w", err)
    }

    if rateType == "" {
        return rates, nil
    }

    filtered := make([]domain.Rate, 0, len(rates))
    for _, r := range rates {
        if strings.EqualFold(r.RateType, rateType) {
            filtered = append(filtered, r)
        }
    }
    return filtered, nil
}

// GetHistoryRates delegates to the repository with all filter parameters.
func (uc *RateUsecase) GetHistoryRates(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
    rates, err := uc.repo.GetHistoryRates(ctx, currency, rateType, from, to, limit)
    if err != nil {
        slog.Error("get history rates failed", "method", "GetHistoryRates", "error", err)
        return nil, fmt.Errorf("GetHistoryRates: %w", err)
    }
    return rates, nil
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `ScrapeRates` success path | Mock scraper returns 2 rates, mock repo records call. Assert count=2, err=nil. |
| Unit | `ScrapeRates` scraper error | Mock scraper returns error. Assert SaveRates never called, error wrapped. |
| Unit | `ScrapeRates` repo error | Mock scraper returns rates, mock repo errors. Assert error logged + wrapped. |
| Unit | `GetCurrentRates` no filter | Mock repo returns 3 rates. Assert all returned. |
| Unit | `GetCurrentRates` with filter | Mock repo returns 3 rates (2 reference, 1 buy). Filter "reference" → assert 2. |
| Unit | `GetCurrentRates` empty result | Mock repo returns empty slice. Assert empty slice, nil error. |
| Unit | `GetHistoryRates` delegation | Mock repo returns rates. Assert all params forwarded, result returned. |
| Unit | `GetHistoryRates` repo error | Mock repo errors. Assert error logged + wrapped. |

### Mock Structures

```go
type mockScraper struct {
    rates []domain.Rate
    err   error
}

func (m *mockScraper) Scrape(ctx context.Context) ([]domain.Rate, error) {
    return m.rates, m.err
}

type mockRepository struct {
    rates      []domain.Rate
    saveErr    error
    latestErr  error
    historyErr error
    savedRates []domain.Rate // captured for assertions
}

func (m *mockRepository) SaveRates(ctx context.Context, rates []domain.Rate) error {
    m.savedRates = rates
    return m.saveErr
}

func (m *mockRepository) GetLatestRates(ctx context.Context, currency string) ([]domain.Rate, error) {
    return m.rates, m.latestErr
}

func (m *mockRepository) GetHistoryRates(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
    return m.rates, m.historyErr
}
```

## Migration / Rollout

No migration required. Purely additive — new package, no changes to existing code. `cmd/api/main.go` will be wired in a future change.

## Open Questions

None — all decisions are clear from the spec, existing interfaces, and project conventions.
