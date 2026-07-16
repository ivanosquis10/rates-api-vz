package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

func TestNewInMemoryDB(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("expected non-nil *sql.DB")
	}

	if err := db.Ping(); err != nil {
		t.Errorf("expected pingable DB, got error: %v", err)
	}
}

func TestNewInvalidPath(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestMigrationIdempotency(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		t.Errorf("second migrate() call failed: %v", err)
	}
}

func TestRatesTableExists(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	var tableName string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='rates'",
	).Scan(&tableName)
	if err != nil {
		t.Fatalf("rates table not found: %v", err)
	}
	if tableName != "rates" {
		t.Errorf("expected table name 'rates', got %q", tableName)
	}
}

func TestRatesTableColumns(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("PRAGMA table_info(rates)")
	if err != nil {
		t.Fatalf("PRAGMA table_info failed: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		columns[name] = true
	}

	expected := []string{"id", "currency", "value", "scraped_at"}
	for _, col := range expected {
		if !columns[col] {
			t.Errorf("missing expected column %q", col)
		}
	}

	unexpected := []string{"rate_type", "bank"}
	for _, col := range unexpected {
		if columns[col] {
			t.Errorf("unexpected column %q found", col)
		}
	}
}

func TestUniqueConstraintEnforced(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rate := domain.Rate{
		Currency:  "USD",
		Value:     36.50,
		ScrapedAt: now,
	}

	err = insertRate(db, rate)
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	// Insert duplicate using insertRate (INSERT OR REPLACE) -> succeeds
	err = insertRate(db, rate)
	if err != nil {
		t.Errorf("duplicate insert failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM rates").Scan(&count)
	if err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row after replace, got %d", count)
	}
}

func TestSaveAndGetLatestRate(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		{Currency: "USD", Value: 36.50, ScrapedAt: now},
	}

	ctx := context.Background()
	s := NewStore(db)
	err = s.SaveRates(ctx, rates)
	if err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	latest, err := s.GetLatestRate(ctx, "USD")
	if err != nil {
		t.Fatalf("GetLatestRate failed: %v", err)
	}

	if latest.Currency != "USD" {
		t.Errorf("expected currency=USD, got %s", latest.Currency)
	}
	if latest.Value != 36.50 {
		t.Errorf("expected value=36.50, got %f", latest.Value)
	}
}

func TestGetLatestRateNotFound(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	s := NewStore(db)

	_, err = s.GetLatestRate(ctx, "EUR")
	if err == nil {
		t.Fatal("expected error for nonexistent rate, got nil")
	}
}

func TestGetHistoryRates(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		{Currency: "USD", Value: 36.00, ScrapedAt: now.Add(-2 * time.Hour)},
		{Currency: "USD", Value: 36.50, ScrapedAt: now},
		{Currency: "EUR", Value: 42.00, ScrapedAt: now},
	}

	ctx := context.Background()
	s := NewStore(db)
	err = s.SaveRates(ctx, rates)
	if err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	history, err := s.GetHistoryRates(ctx, "USD", "", "", 10)
	if err != nil {
		t.Fatalf("GetHistoryRates failed: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 USD rates, got %d", len(history))
	}

	if history[0].ScrapedAt.Before(history[1].ScrapedAt) {
		t.Error("expected rates ordered by scraped_at DESC")
	}
}

func TestGetHistoryRatesWithLimit(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		{Currency: "USD", Value: 36.00, ScrapedAt: now.Add(-3 * time.Hour)},
		{Currency: "USD", Value: 36.50, ScrapedAt: now.Add(-2 * time.Hour)},
		{Currency: "USD", Value: 37.00, ScrapedAt: now.Add(-1 * time.Hour)},
	}

	ctx := context.Background()
	s := NewStore(db)
	err = s.SaveRates(ctx, rates)
	if err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	history, err := s.GetHistoryRates(ctx, "USD", "", "", 2)
	if err != nil {
		t.Fatalf("GetHistoryRates failed: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 rates with limit=2, got %d", len(history))
	}
}

func TestInterfaceSatisfaction(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	var _ domain.Repository = NewStore(db)
}

func TestSaveRatesDuplicateSkipped(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	s := NewStore(db)

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rate1 := domain.Rate{Currency: "USD", Value: 36.50, ScrapedAt: now}
	rate2 := domain.Rate{Currency: "USD", Value: 37.00, ScrapedAt: now} // duplicate key, different value

	// Save rate1
	if err := s.SaveRates(ctx, []domain.Rate{rate1}); err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	// Try to save rate2, should be skipped
	if err := s.SaveRates(ctx, []domain.Rate{rate2}); err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	// Value should still be 36.50 (skipped rate2 insertion)
	latest, err := s.GetLatestRate(ctx, "USD")
	if err != nil {
		t.Fatalf("GetLatestRate failed: %v", err)
	}
	if latest.Value != 36.50 {
		t.Errorf("expected value to remain 36.50, got %f", latest.Value)
	}
}

func TestMigrationRecreation(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	legacySchema := `
	CREATE TABLE rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		currency TEXT NOT NULL,
		rate_type TEXT NOT NULL,
		bank TEXT,
		value REAL NOT NULL,
		scraped_at DATETIME NOT NULL
	);`
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatalf("failed to create legacy table: %v", err)
	}

	if err := migrate(db); err != nil {
		t.Fatalf("migrate failed on legacy table: %v", err)
	}

	rows, err := db.Query("PRAGMA table_info(rates)")
	if err != nil {
		t.Fatalf("PRAGMA table_info failed: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		columns[name] = true
	}

	if columns["rate_type"] || columns["bank"] {
		t.Error("legacy columns rate_type or bank were not dropped")
	}
	if !columns["id"] || !columns["currency"] || !columns["value"] || !columns["scraped_at"] {
		t.Error("missing new table columns")
	}
}
