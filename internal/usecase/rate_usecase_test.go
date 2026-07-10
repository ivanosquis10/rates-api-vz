package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

// mockScraper implements scraper.Scraper for testing.
type mockScraper struct {
	rates []domain.Rate
	err   error
}

func (m *mockScraper) Scrape(ctx context.Context) ([]domain.Rate, error) {
	return m.rates, m.err
}

// mockRepository implements domain.Repository for testing.
type mockRepository struct {
	rates      []domain.Rate
	saveErr    error
	latestErr  error
	historyErr error
	savedRates []domain.Rate
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

// --- ScrapeRates tests ---

func TestScrapeRates_Success(t *testing.T) {
	rates := []domain.Rate{
		{Currency: "USD", RateType: "reference", Value: 36.5},
		{Currency: "EUR", RateType: "reference", Value: 40.0},
	}

	scraper := &mockScraper{rates: rates, err: nil}
	repo := &mockRepository{}

	uc := NewRateUsecase(repo, scraper)

	count, err := uc.ScrapeRates(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(repo.savedRates) != 2 {
		t.Errorf("expected SaveRates to be called with 2 rates, got %d", len(repo.savedRates))
	}
}

func TestScrapeRates_ScraperError(t *testing.T) {
	scraper := &mockScraper{rates: nil, err: domain.ErrScrapeFailed}
	repo := &mockRepository{}

	uc := NewRateUsecase(repo, scraper)

	count, err := uc.ScrapeRates(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
	if repo.savedRates != nil {
		t.Error("expected SaveRates not to be called")
	}
}

func TestScrapeRates_RepoError(t *testing.T) {
	rates := []domain.Rate{
		{Currency: "USD", RateType: "reference", Value: 36.5},
	}

	scraper := &mockScraper{rates: rates, err: nil}
	repo := &mockRepository{saveErr: domain.ErrDatabase}

	uc := NewRateUsecase(repo, scraper)

	count, err := uc.ScrapeRates(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
	if len(repo.savedRates) != 1 {
		t.Errorf("expected SaveRates to be called with 1 rate, got %d", len(repo.savedRates))
	}
}

// --- GetCurrentRates tests ---

func TestGetCurrentRates_NoFilter(t *testing.T) {
	now := time.Now()
	rates := []domain.Rate{
		{Currency: "USD", RateType: "reference", Value: 36.5, ScrapedAt: now},
		{Currency: "USD", RateType: "buy", Value: 35.0, ScrapedAt: now},
		{Currency: "USD", RateType: "sell", Value: 37.0, ScrapedAt: now},
	}

	scraper := &mockScraper{}
	repo := &mockRepository{rates: rates}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetCurrentRates(context.Background(), "USD", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 rates, got %d", len(result))
	}
}

func TestGetCurrentRates_WithFilter(t *testing.T) {
	now := time.Now()
	rates := []domain.Rate{
		{Currency: "USD", RateType: "reference", Value: 36.5, ScrapedAt: now},
		{Currency: "USD", RateType: "reference", Value: 36.6, ScrapedAt: now},
		{Currency: "USD", RateType: "buy", Value: 35.0, ScrapedAt: now},
	}

	scraper := &mockScraper{}
	repo := &mockRepository{rates: rates}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetCurrentRates(context.Background(), "USD", "reference")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 reference rates, got %d", len(result))
	}
	for _, r := range result {
		if r.RateType != "reference" {
			t.Errorf("expected rate_type 'reference', got '%s'", r.RateType)
		}
	}
}

func TestGetCurrentRates_EmptyResult(t *testing.T) {
	scraper := &mockScraper{}
	repo := &mockRepository{rates: []domain.Rate{}}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetCurrentRates(context.Background(), "XYZ", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 rates, got %d", len(result))
	}
}

// --- GetHistoryRates tests ---

func TestGetHistoryRates_Delegation(t *testing.T) {
	now := time.Now()
	rates := []domain.Rate{
		{Currency: "USD", RateType: "buy", Value: 35.0, ScrapedAt: now},
	}

	scraper := &mockScraper{}
	repo := &mockRepository{rates: rates}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetHistoryRates(context.Background(), "USD", "buy", "2026-01-01", "2026-07-01", 50)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 rate, got %d", len(result))
	}
	if result[0].RateType != "buy" {
		t.Errorf("expected rate_type 'buy', got '%s'", result[0].RateType)
	}
}

func TestGetHistoryRates_RepoError(t *testing.T) {
	scraper := &mockScraper{}
	repo := &mockRepository{historyErr: domain.ErrNotFound}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetHistoryRates(context.Background(), "USD", "buy", "2026-01-01", "2026-07-01", 50)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
