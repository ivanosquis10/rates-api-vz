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

// ─── Test Helpers ───────────────────────────────────────────────────────────

// newTestStore creates an in-memory Store with automatic cleanup.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("newTestStore: New(:memory:) failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStore(db)
}

// createTestRate returns a domain.Rate with sensible defaults for the given
// currency, rateType, bank, and scrapedAt. Value is derived from a deterministic
// hash so different inputs produce different values.
func createTestRate(currency, rateType, bank string, scrapedAt time.Time, value float64) domain.Rate {
	return domain.Rate{
		Currency:  currency,
		RateType:  rateType,
		Bank:      bank,
		Value:     value,
		ScrapedAt: scrapedAt,
	}
}

// ─── Phase 2: SaveRates Tests ──────────────────────────────────────────────

func TestSaveRatesBankSpecific(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		createTestRate("USD", "reference", "", now, 36.50),
		createTestRate("USD", "reference", "Banco de Venezuela", now.Add(time.Minute), 37.00),
	}

	if err := s.SaveRates(ctx, rates); err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	// Verify both rows persisted with correct bank values via raw query.
	db := s.db
	rows, err := db.QueryContext(ctx,
		`SELECT currency, rate_type, bank, value FROM rates ORDER BY scraped_at`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var got []struct {
		currency, rateType, bank string
		value                    float64
	}
	for rows.Next() {
		var r struct {
			currency, rateType, bank string
			value                    float64
		}
		if err := rows.Scan(&r.currency, &r.rateType, &r.bank, &r.value); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		got = append(got, r)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}

	// Reference rate: bank should be empty string
	if got[0].bank != "" {
		t.Errorf("reference rate bank: got %q, want empty string", got[0].bank)
	}
	if got[0].value != 36.50 {
		t.Errorf("reference rate value: got %v, want 36.50", got[0].value)
	}

	// Bank-specific rate: bank should be "Banco de Venezuela"
	if got[1].bank != "Banco de Venezuela" {
		t.Errorf("bank-specific rate bank: got %q, want %q", got[1].bank, "Banco de Venezuela")
	}
	if got[1].value != 37.00 {
		t.Errorf("bank-specific rate value: got %v, want 37.00", got[1].value)
	}
}

func TestSaveRatesDuplicateViaAPI(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rate := createTestRate("USD", "reference", "", now, 36.50)

	// First save succeeds
	if err := s.SaveRates(ctx, []domain.Rate{rate}); err != nil {
		t.Fatalf("first SaveRates failed: %v", err)
	}

	// Second save with identical (currency, rate_type, bank, scraped_at) must fail
	err := s.SaveRates(ctx, []domain.Rate{rate})
	if err == nil {
		t.Error("expected error on duplicate SaveRates, got nil")
	}

	// Verify original row unchanged
	latest, err := s.GetLatestRates(ctx, "USD")
	if err != nil {
		t.Fatalf("GetLatestRates failed: %v", err)
	}
	if len(latest) != 1 {
		t.Fatalf("expected 1 rate after failed duplicate, got %d", len(latest))
	}
	if latest[0].Value != 36.50 {
		t.Errorf("original rate value changed: got %v, want 36.50", latest[0].Value)
	}
}

// ─── Phase 3: GetLatestRates Tests ─────────────────────────────────────────

func TestGetLatestRatesMultiTimestamp(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	base := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		createTestRate("USD", "reference", "", base.Add(10*time.Hour), 36.00),  // 10:00
		createTestRate("USD", "reference", "", base.Add(12*time.Hour), 36.50),  // 12:00 — should win
	}

	if err := s.SaveRates(ctx, rates); err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	latest, err := s.GetLatestRates(ctx, "USD")
	if err != nil {
		t.Fatalf("GetLatestRates failed: %v", err)
	}
	if len(latest) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(latest))
	}
	if latest[0].Value != 36.50 {
		t.Errorf("expected latest value 36.50, got %v", latest[0].Value)
	}
	wantTime := base.Add(12 * time.Hour)
	if !latest[0].ScrapedAt.Equal(wantTime) {
		t.Errorf("expected scraped_at %v, got %v", wantTime, latest[0].ScrapedAt)
	}
}

func TestGetLatestRatesMultiBank(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	base := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		// Banco A: 10:00 and 12:00
		createTestRate("USD", "reference", "Banco A", base.Add(10*time.Hour), 36.00),
		createTestRate("USD", "reference", "Banco A", base.Add(12*time.Hour), 37.00),
		// Banco B: 11:00
		createTestRate("USD", "reference", "Banco B", base.Add(11*time.Hour), 36.50),
	}

	if err := s.SaveRates(ctx, rates); err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	latest, err := s.GetLatestRates(ctx, "USD")
	if err != nil {
		t.Fatalf("GetLatestRates failed: %v", err)
	}

	// GROUP BY currency, rate_type → returns latest per (currency, rate_type) pair.
	// All three rates have rate_type="reference", so only 1 row returned (latest of all).
	if len(latest) != 1 {
		t.Fatalf("expected 1 rate (latest per rate_type), got %d", len(latest))
	}

	// The latest across all banks for rate_type=reference is Banco A at 12:00
	if latest[0].Bank != "Banco A" {
		t.Errorf("expected bank %q, got %q", "Banco A", latest[0].Bank)
	}
	if latest[0].Value != 37.00 {
		t.Errorf("expected value 37.00, got %v", latest[0].Value)
	}
}

// ─── Phase 4: GetHistoryRates Tests ────────────────────────────────────────

func TestGetHistoryRatesTableDriven(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		// USD reference on 3 dates
		createTestRate("USD", "reference", "", base.Add(0), 36.00),      // Jul 1
		createTestRate("USD", "reference", "", base.Add(4*24*time.Hour), 36.50), // Jul 5
		createTestRate("USD", "reference", "", base.Add(9*24*time.Hour), 37.00), // Jul 10
		// USD parallel on same dates
		createTestRate("USD", "parallel", "", base.Add(0), 100.00),
		createTestRate("USD", "parallel", "", base.Add(4*24*time.Hour), 101.00),
		createTestRate("USD", "parallel", "", base.Add(9*24*time.Hour), 102.00),
		// EUR reference on Jul 1 only
		createTestRate("EUR", "reference", "", base.Add(0), 42.00),
	}

	if err := s.SaveRates(ctx, rates); err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	cases := []struct {
		name      string
		currency  string
		rateType  string
		from      string
		to        string
		limit     int
		wantCount int
	}{
		{"rateType filter only", "USD", "reference", "", "", 100, 3},
		{"date range filter only", "USD", "", "2026-07-03", "2026-07-08", 100, 2},
		{"combined filters", "USD", "reference", "2026-07-01", "2026-07-10", 30, 2},
		{"no match", "XYZ", "", "", "", 100, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := s.GetHistoryRates(ctx, tc.currency, tc.rateType, tc.from, tc.to, tc.limit)
			if err != nil {
				t.Fatalf("GetHistoryRates failed: %v", err)
			}
			if len(result) != tc.wantCount {
				t.Errorf("expected %d rates, got %d", tc.wantCount, len(result))
			}

			// Verify ordering is DESC
			for i := 1; i < len(result); i++ {
				if result[i-1].ScrapedAt.Before(result[i].ScrapedAt) {
					t.Errorf("results not in DESC order at index %d", i)
				}
			}

			// Verify filter correctness for non-empty results
			if tc.wantCount > 0 && tc.rateType != "" {
				for _, r := range result {
					if r.RateType != tc.rateType {
						t.Errorf("got rate_type %q, want %q", r.RateType, tc.rateType)
					}
				}
			}
		})
	}
}

func TestGetHistoryRatesOrdering(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rates := []domain.Rate{
		createTestRate("USD", "reference", "", now.Add(-2*time.Hour), 36.00),
		createTestRate("USD", "reference", "", now, 36.50),
		createTestRate("USD", "reference", "", now.Add(-1*time.Hour), 36.25),
	}

	if err := s.SaveRates(ctx, rates); err != nil {
		t.Fatalf("SaveRates failed: %v", err)
	}

	history, err := s.GetHistoryRates(ctx, "USD", "", "", "", 10)
	if err != nil {
		t.Fatalf("GetHistoryRates failed: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 rates, got %d", len(history))
	}

	// Must be ordered by scraped_at DESC (most recent first)
	expectedOrder := []float64{36.50, 36.25, 36.00}
	for i, wantVal := range expectedOrder {
		if history[i].Value != wantVal {
			t.Errorf("index %d: expected value %v, got %v", i, wantVal, history[i].Value)
		}
	}

	// Verify strict descending timestamps
	for i := 1; i < len(history); i++ {
		if !history[i-1].ScrapedAt.After(history[i].ScrapedAt) &&
			!history[i-1].ScrapedAt.Equal(history[i].ScrapedAt) {
			t.Errorf("not DESC: history[%d].ScrapedAt=%v not after history[%d].ScrapedAt=%v",
				i-1, history[i-1].ScrapedAt, i, history[i].ScrapedAt)
		}
	}
}

func TestGetHistoryRatesNilSafety(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Query nonexistent currency on empty store
	result, err := s.GetHistoryRates(ctx, "NONEXISTENT", "", "", "", 100)
	if err != nil {
		t.Fatalf("GetHistoryRates failed: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil slice for empty result")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 rates, got %d", len(result))
	}
}
