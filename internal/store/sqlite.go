package store

import (
	"context"
	"database/sql"
	"errors"
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
	// Check if the legacy columns rate_type or bank exist
	rows, err := db.Query("PRAGMA table_info(rates)")
	if err == nil {
		defer rows.Close()
		shouldDrop := false
		for rows.Next() {
			var cid int
			var name string
			var typeVal string
			var notnull int
			var dfltVal sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &typeVal, &notnull, &dfltVal, &pk); err == nil {
				if name == "rate_type" || name == "bank" {
					shouldDrop = true
				}
			}
		}
		if shouldDrop {
			if _, err := db.Exec("DROP TABLE IF EXISTS rates"); err != nil {
				return fmt.Errorf("drop legacy rates table: %w", err)
			}
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		currency TEXT NOT NULL,
		value REAL NOT NULL,
		scraped_at DATETIME NOT NULL,
		UNIQUE(currency, scraped_at)
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
		`INSERT OR REPLACE INTO rates (currency, value, scraped_at)
		 VALUES (?, ?, ?)`,
		r.Currency, r.Value, r.ScrapedAt,
	)
	return err
}

// SaveRates persists a batch of rates atomically, skipping duplicates.
func (s *Store) SaveRates(ctx context.Context, rates []domain.Rate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	checkStmt, err := tx.PrepareContext(ctx, `SELECT 1 FROM rates WHERE currency = ? AND scraped_at = ? LIMIT 1`)
	if err != nil {
		return fmt.Errorf("prepare check statement: %w", err)
	}
	defer checkStmt.Close()

	insertStmt, err := tx.PrepareContext(ctx,
		`INSERT INTO rates (currency, value, scraped_at)
		 VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	for _, r := range rates {
		var exists int
		err := checkStmt.QueryRowContext(ctx, r.Currency, r.ScrapedAt).Scan(&exists)
		if err == nil {
			// Already exists, skip
			continue
		} else if err != sql.ErrNoRows {
			return fmt.Errorf("check rate existence: %w", err)
		}

		if _, err := insertStmt.ExecContext(ctx, r.Currency, r.Value, r.ScrapedAt); err != nil {
			return fmt.Errorf("insert rate: %w", err)
		}
	}

	return tx.Commit()
}

// GetLatestRate returns the most recent rate for the given currency.
func (s *Store) GetLatestRate(ctx context.Context, currency string) (domain.Rate, error) {
	var r domain.Rate
	err := s.db.QueryRowContext(ctx,
		`SELECT id, currency, value, scraped_at
		 FROM rates
		 WHERE currency = ?
		 ORDER BY scraped_at DESC
		 LIMIT 1`,
		currency).Scan(&r.ID, &r.Currency, &r.Value, &r.ScrapedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Rate{}, domain.ErrNotFound
		}
		return domain.Rate{}, err
	}
	return r, nil
}

// GetHistoryRates returns rates matching the given filters, ordered by
// scraped_at DESC, limited to the specified count. Returns empty slice
// (not nil) when no rates match.
func (s *Store) GetHistoryRates(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error) {
	query := `SELECT id, currency, value, scraped_at
		FROM rates
		WHERE currency = ?`
	args := []any{currency}

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
		if err := rows.Scan(&r.ID, &r.Currency, &r.Value, &r.ScrapedAt); err != nil {
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
