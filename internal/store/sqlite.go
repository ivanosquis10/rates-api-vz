package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	_ "modernc.org/sqlite"
)

// New opens a SQLite database at dbPath and runs the schema migration.
// Use ":memory:" for an in-memory database (useful for testing).
// Returns an error if dbPath is empty.
func New(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path must not be empty")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}

// migrate runs idempotent schema creation for the rates table.
func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		currency TEXT NOT NULL,
		rate_type TEXT NOT NULL,
		bank TEXT,
		value REAL NOT NULL,
		scraped_at DATETIME NOT NULL,
		UNIQUE(currency, rate_type, bank, scraped_at)
	);

	CREATE INDEX IF NOT EXISTS idx_rates_currency_scraped
		ON rates(currency, scraped_at DESC);

	CREATE INDEX IF NOT EXISTS idx_rates_scraped
		ON rates(scraped_at DESC);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}

// Store implements domain.Repository using SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates a Store backed by the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// insertRate inserts a single rate row (used by tests and internally).
func insertRate(db *sql.DB, r domain.Rate) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO rates (currency, rate_type, bank, value, scraped_at)
		 VALUES (?, ?, ?, ?, ?)`,
		r.Currency, r.RateType, r.Bank, r.Value, r.ScrapedAt,
	)
	return err
}

// SaveRates persists a batch of rates atomically.
func (s *Store) SaveRates(ctx context.Context, rates []domain.Rate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO rates (currency, rate_type, bank, value, scraped_at)
		 VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range rates {
		if _, err := stmt.ExecContext(ctx, r.Currency, r.RateType, r.Bank, r.Value, r.ScrapedAt); err != nil {
			return fmt.Errorf("insert rate: %w", err)
		}
	}

	return tx.Commit()
}

// GetLatestRates returns the most recent rate for each (currency, rate_type)
// combination for the given currency. Returns empty slice (not nil) when no
// rates exist.
func (s *Store) GetLatestRates(ctx context.Context, currency string) ([]domain.Rate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, currency, rate_type, bank, value, scraped_at
		 FROM rates
		 WHERE ? = '' OR currency = ?
		 GROUP BY currency, rate_type
		 HAVING scraped_at = MAX(scraped_at)`,
		currency, currency)
	if err != nil {
		return nil, fmt.Errorf("query latest rates: %w", err)
	}
	defer rows.Close()

	return scanRates(rows)
}

// GetHistoryRates returns rates matching the given filters, ordered by
// scraped_at DESC, limited to the specified count. Returns empty slice
// (not nil) when no rates match.
func (s *Store) GetHistoryRates(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
	query := `SELECT id, currency, rate_type, bank, value, scraped_at
		FROM rates
		WHERE currency = ?`
	args := []any{currency}

	if rateType != "" {
		query += ` AND rate_type = ?`
		args = append(args, rateType)
	}
	if from != "" {
		query += ` AND scraped_at >= ?`
		args = append(args, from)
	}
	if to != "" {
		query += ` AND scraped_at <= ?`
		args = append(args, to)
	}

	query += ` ORDER BY scraped_at DESC`

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query history rates: %w", err)
	}
	defer rows.Close()

	return scanRates(rows)
}

// scanRates reads all rows from a *sql.Rows into a slice of domain.Rate.
func scanRates(rows *sql.Rows) ([]domain.Rate, error) {
	var rates []domain.Rate
	for rows.Next() {
		var r domain.Rate
		if err := rows.Scan(&r.ID, &r.Currency, &r.RateType, &r.Bank, &r.Value, &r.ScrapedAt); err != nil {
			return nil, fmt.Errorf("scan rate: %w", err)
		}
		rates = append(rates, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	if rates == nil {
		rates = []domain.Rate{}
	}
	return rates, nil
}
