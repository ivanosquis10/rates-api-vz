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
	rate       domain.Rate
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

func (m *mockRepository) GetLatestRate(ctx context.Context, currency string) (domain.Rate, error) {
	return m.rate, m.latestErr
}

func (m *mockRepository) GetHistoryRates(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error) {
	return m.rates, m.historyErr
}

// --- ScrapeRates tests ---

func TestScrapeRates_Success(t *testing.T) {
	rates := []domain.Rate{
		{Currency: "USD", Value: 36.5},
		{Currency: "EUR", Value: 40.0},
	}

	scraper := &mockScraper{rates: rates, err: nil}
	repo := &mockRepository{}

	uc := NewRateUsecase(repo, scraper)

	scrapedRates, err := uc.ScrapeRates(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scrapedRates) != 2 {
		t.Errorf("expected 2 rates, got %d", len(scrapedRates))
	}
	if len(repo.savedRates) != 2 {
		t.Errorf("expected SaveRates to be called with 2 rates, got %d", len(repo.savedRates))
	}
}

func TestScrapeRates_ScraperError(t *testing.T) {
	scraper := &mockScraper{rates: nil, err: domain.ErrScrapeFailed}
	repo := &mockRepository{}

	uc := NewRateUsecase(repo, scraper)

	scrapedRates, err := uc.ScrapeRates(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if scrapedRates != nil {
		t.Errorf("expected nil rates, got %v", scrapedRates)
	}
	if repo.savedRates != nil {
		t.Error("expected SaveRates not to be called")
	}
}

func TestScrapeRates_RepoError(t *testing.T) {
	rates := []domain.Rate{
		{Currency: "USD", Value: 36.5},
	}

	scraper := &mockScraper{rates: rates, err: nil}
	repo := &mockRepository{saveErr: domain.ErrDatabase}

	uc := NewRateUsecase(repo, scraper)

	scrapedRates, err := uc.ScrapeRates(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if scrapedRates != nil {
		t.Errorf("expected nil rates, got %v", scrapedRates)
	}
	if len(repo.savedRates) != 1 {
		t.Errorf("expected SaveRates to be called with 1 rate, got %d", len(repo.savedRates))
	}
}

// --- GetLatestRate tests ---

func TestGetLatestRate_Success(t *testing.T) {
	now := time.Now()
	rate := domain.Rate{Currency: "USD", Value: 36.5, ScrapedAt: now}

	scraper := &mockScraper{}
	repo := &mockRepository{rate: rate}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetLatestRate(context.Background(), "USD")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Currency != "USD" {
		t.Errorf("expected USD, got %s", result.Currency)
	}
	if result.Value != 36.5 {
		t.Errorf("expected 36.5, got %f", result.Value)
	}
}

func TestGetLatestRate_Error(t *testing.T) {
	scraper := &mockScraper{}
	repo := &mockRepository{latestErr: domain.ErrNotFound}

	uc := NewRateUsecase(repo, scraper)

	_, err := uc.GetLatestRate(context.Background(), "EUR")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetHistoryRates tests ---

func TestGetHistoryRates_Success(t *testing.T) {
	now := time.Now()
	rates := []domain.Rate{
		{Currency: "USD", Value: 35.0, ScrapedAt: now},
	}

	scraper := &mockScraper{}
	repo := &mockRepository{rates: rates}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetHistoryRates(context.Background(), "USD", "2026-01-01", "2026-07-01", 50)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 rate, got %d", len(result))
	}
	if result[0].Value != 35.0 {
		t.Errorf("expected value 35.0, got %f", result[0].Value)
	}
}

func TestGetHistoryRates_Error(t *testing.T) {
	scraper := &mockScraper{}
	repo := &mockRepository{historyErr: domain.ErrDatabase}

	uc := NewRateUsecase(repo, scraper)

	result, err := uc.GetHistoryRates(context.Background(), "USD", "2026-01-01", "2026-07-01", 50)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
