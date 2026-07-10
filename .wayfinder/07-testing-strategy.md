# Ticket 07: Testing Strategy

**Type**: Grilling (HITL)
**Blocked by**: Ticket 03 (architecture structure)
**Status**: RESOLVED

## Question

What testing approach and patterns should we use?

## Resolution

### Approach: Manual mocks + table-driven tests + SQLite in-memory

### Testing layers

```
┌─────────────────────────────────────────┐
│           handler tests                 │  Mock usecase → test HTTP responses
│           (unit)                        │
├─────────────────────────────────────────┤
│           usecase tests                 │  Mock repository + scraper → test logic
│           (unit)                        │
├─────────────────────────────────────────┤
│           repository tests              │  SQLite in-memory → test SQL queries
│           (integration-lite)            │
└─────────────────────────────────────────┘
```

### Manual mocks

```go
// usecase/mock_test.go
package usecase_test

type mockRepository struct {
    rates   []domain.Rate
    err     error
}

func (m *mockRepository) SaveRates(ctx context.Context, rates []domain.Rate) error {
    return m.err
}

func (m *mockRepository) GetLatestRates(ctx context.Context, currency string) ([]domain.Rate, error) {
    return m.rates, m.err
}

// Similar for scraper mock...
```

**Why manual mocks:**
- No external dependencies (mockery, gomock, testify)
- Full control over behavior
- Simple to understand and maintain
- Interfaces are small (1-3 methods each)

### Table-driven tests

```go
// usecase/rate_usecase_test.go
func TestGetCurrentRates(t *testing.T) {
    tests := []struct {
        name      string
        repo      *mockRepository
        want      []domain.Rate
        wantErr   bool
    }{
        {
            name: "returns rates successfully",
            repo: &mockRepository{
                rates: []domain.Rate{
                    {Currency: "USD", RateType: "reference", Value: 709.69},
                },
            },
            want: []domain.Rate{
                {Currency: "USD", RateType: "reference", Value: 709.69},
            },
        },
        {
            name:    "returns error when repo fails",
            repo:    &mockRepository{err: errors.New("db error")},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            uc := NewRateUseCase(tt.repo, nil)
            got, err := uc.GetCurrentRates(context.Background())
            
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Repository tests (SQLite in-memory)

```go
// repository/sqlite/rate_repo_test.go
func TestSaveAndGetRates(t *testing.T) {
    db, _ := sql.Open("sqlite", ":memory:")
    defer db.Close()
    
    // Run migrations
    migrate(db)
    
    repo := NewSQLiteRateRepository(db)
    
    // Save rates
    rates := []domain.Rate{
        {Currency: "USD", RateType: "reference", Value: 709.69, ScrapedAt: time.Now()},
    }
    err := repo.SaveRates(context.Background(), rates)
    assert.NoError(t, err)
    
    // Get latest
    got, err := repo.GetLatestRates(context.Background(), "USD")
    assert.NoError(t, err)
    assert.Len(t, got, 1)
    assert.Equal(t, 709.69, got[0].Value)
}
```

### Test file organization

```
internal/
├── usecase/
│   ├── rate_usecase.go
│   ├── rate_usecase_test.go      # Unit tests
│   └── mock_test.go              # Mocks (in test package)
├── repository/
│   └── sqlite/
│       ├── rate_repo.go
│       └── rate_repo_test.go     # SQLite in-memory tests
├── handler/
│   ├── rate_handler.go
│   ├── rate_handler_test.go      # HTTP handler tests
│   └── mock_test.go              # Mocks
```

### What to test

| Package | What to test | How |
|---------|-------------|-----|
| **usecase** | Business logic, error cases, validation | Manual mocks of repository + scraper |
| **repository** | SQL queries, constraints, edge cases | SQLite `:memory:` |
| **handler** | HTTP status codes, JSON responses, auth | Manual mocks of usecase |
| **scraper** | HTML parsing, value extraction | Mock HTTP responses (httptest) |
| **middleware** | Auth validation, rate limiting | Direct function calls |

### Testing conventions

1. **Table-driven tests** — standard Go pattern, one `tests` slice per function
2. **`_test.go` alongside source** — not separate `tests/` directory
3. **`t.Helper()`** — mark helper functions
4. **`t.Parallel()`** — where safe
5. **No testify** — use stdlib `testing` + `reflect.DeepEqual` + simple assertions
6. **Test names**: `TestFunctionName/scenario_description`

### Coverage target

- **Use cases**: 90%+ (business logic must be thoroughly tested)
- **Repository**: 80%+ (SQL queries, edge cases)
- **Handler**: 80%+ (HTTP responses, error mapping)
- **Scraper**: 70%+ (HTML parsing, mocked responses)
