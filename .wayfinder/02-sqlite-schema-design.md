# Ticket 02: SQLite Schema Design

**Type**: Grilling (HITL)
**Blocked by**: None (frontier)
**Status**: RESOLVED

## Question

What should the SQLite schema look like for storing exchange rates?

## Resolution

### Design decisions

1. **Single table** — `rates` handles both BCV reference and bank rates
2. **Multiple entries per day** — captures retry attempts, gives full audit trail
3. **Bank is nullable** — NULL for BCV reference rates, bank name for bank rates
4. **Unique constraint** — prevents exact duplicate entries
5. **Index on查询patterns** — currency + date for fast lookups

### Schema

```sql
CREATE TABLE rates (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    currency    TEXT    NOT NULL,                    -- 'USD' or 'EUR'
    rate_type   TEXT    NOT NULL,                    -- 'reference' | 'buy' | 'sell'
    bank        TEXT,                                -- NULL for reference, bank name for buy/sell
    value       REAL    NOT NULL,                    -- the rate value (Bs per unit)
    scraped_at  TEXT    NOT NULL,                    -- ISO 8601 datetime of scraping
    source_url  TEXT,                                -- URL scraped from
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Prevent exact duplicate scrapes
CREATE UNIQUE INDEX idx_rates_unique 
    ON rates(currency, rate_type, bank, scraped_at);

-- Fast lookups: "get latest USD rates" or "get history for EUR"
CREATE INDEX idx_rates_currency_date 
    ON rates(currency, scraped_at DESC);

-- Fast lookups: "get all bank rates for a date"
CREATE INDEX idx_rates_date 
    ON rates(scraped_at DESC);
```

### Query examples

```sql
-- Latest reference rates (USD + EUR)
SELECT currency, value, scraped_at 
FROM rates 
WHERE rate_type = 'reference' 
  AND bank IS NULL
GROUP BY currency 
HAVING scraped_at = MAX(scraped_at);

-- All rates for a specific day
SELECT * FROM rates 
WHERE date(scraped_at) = '2026-07-10'
ORDER BY currency, rate_type, bank;

-- Historical USD reference rates
SELECT value, scraped_at FROM rates 
WHERE currency = 'USD' AND rate_type = 'reference'
ORDER BY scraped_at DESC
LIMIT 30;

-- Bank buy/sell comparison for a day
SELECT bank, 
       MAX(CASE WHEN rate_type = 'buy' THEN value END) as compra,
       MAX(CASE WHEN rate_type = 'sell' THEN value END) as venta
FROM rates 
WHERE currency = 'USD' AND date(scraped_at) = '2026-07-10'
  AND bank IS NOT NULL
GROUP BY bank;
```

### Trade-offs

| Decision | Chosen | Alternative | Why |
|----------|--------|-------------|-----|
| Single table | ✅ | Two tables | Simpler queries, less joins, one scraper writes to one table |
| Multiple/day | ✅ | One/day | Full audit trail, retry visibility, no data loss |
| Bank nullable | ✅ | Separate tables | Clean single-table design, reference rates have no bank |
| TEXT dates | ✅ | INTEGER unix | SQLite has great TEXT date functions, more readable |
| REAL values | ✅ | TEXT | Direct math operations, no parsing needed |
