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
	// An invalid path like empty string or path with null bytes should fail
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

	// Run migration a second time — must not error
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

	// Verify the rates table exists by querying it
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

	// Verify expected columns exist
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

	expected := []string{"id", "currency", "rate_type", "bank", "value", "scraped_at"}
	for _, col := range expected {
		if !columns[col] {
			t.Errorf("missing expected column %q", col)
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
		RateType:  "reference",
		Bank:      "",
		Value:     36.50,
		ScrapedAt: now,
	}

	// Insert first row
	err = insertRate(db, rate)
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	// Insert duplicate — must fail with unique constraint violation
	err = insertRate(db, rate)
	if err == nil {
		t.Error("expected unique constraint violation on duplicate insert, got nil")
	}
}

func TestSaveAndGetLatestRates(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		{Currency: "USD", RateType: "reference", Bank: "", Value: 36.50, ScrapedAt: now},
		{Currency: "USD", RateType: "parallel", Bank: "", Value: 100.00, ScrapedAt: now},
	}

	// SaveRates via the store
	ctx := context.Background()
	s := NewStore(db)
	err = s.SaveRates(ctx, rates)
	if err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	// GetLatestRates
	latest, err := s.GetLatestRates(ctx, "USD")
	if err != nil {
		t.Fatalf("GetLatestRates failed: %v", err)
	}
	if len(latest) == 0 {
		t.Fatal("expected non-empty result from GetLatestRates")
	}

	// Should have 2 entries (reference + parallel)
	if len(latest) != 2 {
		t.Errorf("expected 2 rates, got %d", len(latest))
	}
}

func TestGetLatestRatesEmptyCurrency(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) returned error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	s := NewStore(db)

	latest, err := s.GetLatestRates(ctx, "EUR")
	if err != nil {
		t.Fatalf("GetLatestRates failed: %v", err)
	}
	if latest == nil {
		t.Error("expected empty slice (not nil) for no results")
	}
	if len(latest) != 0 {
		t.Errorf("expected 0 rates for unknown currency, got %d", len(latest))
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
		{Currency: "USD", RateType: "reference", Bank: "", Value: 36.00, ScrapedAt: now.Add(-2 * time.Hour)},
		{Currency: "USD", RateType: "reference", Bank: "", Value: 36.50, ScrapedAt: now},
		{Currency: "EUR", RateType: "reference", Bank: "", Value: 42.00, ScrapedAt: now},
	}

	ctx := context.Background()
	s := NewStore(db)
	err = s.SaveRates(ctx, rates)
	if err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	// Get history for USD only
	history, err := s.GetHistoryRates(ctx, "USD", "", "", "", 10)
	if err != nil {
		t.Fatalf("GetHistoryRates failed: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 USD rates, got %d", len(history))
	}

	// Should be ordered by scraped_at DESC (most recent first)
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
		{Currency: "USD", RateType: "reference", Bank: "", Value: 36.00, ScrapedAt: now.Add(-3 * time.Hour)},
		{Currency: "USD", RateType: "reference", Bank: "", Value: 36.50, ScrapedAt: now.Add(-2 * time.Hour)},
		{Currency: "USD", RateType: "reference", Bank: "", Value: 37.00, ScrapedAt: now.Add(-1 * time.Hour)},
	}

	ctx := context.Background()
	s := NewStore(db)
	err = s.SaveRates(ctx, rates)
	if err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	history, err := s.GetHistoryRates(ctx, "USD", "", "", "", 2)
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

	// Compile-time check: *Store must satisfy domain.Repository
	var _ domain.Repository = NewStore(db)
}
