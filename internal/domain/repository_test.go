package domain

import (
	"context"
	"testing"
	"time"
)

// Verify that a concrete type can satisfy the Repository interface at compile time.
type testRepository struct{}

func (testRepository) SaveRates(_ context.Context, _ []Rate) error                        { return nil }
func (testRepository) GetLatestRates(_ context.Context, _ string) ([]Rate, error)         { return nil, nil }
func (testRepository) GetHistoryRates(_ context.Context, _, _, _, _ string, _ int) ([]Rate, error) {
	return nil, nil
}

func TestRepositoryInterfaceSatisfied(t *testing.T) {
	// Compile-time check: testRepository must implement Repository
	var _ Repository = testRepository{}
}

func TestRepositoryMethodSignatures(t *testing.T) {
	ctx := context.Background()
	repo := testRepository{}

	// SaveRates accepts context and slice of Rate
	err := repo.SaveRates(ctx, []Rate{
		{Currency: "USD", RateType: "reference", Value: 36.5, ScrapedAt: time.Now()},
	})
	if err != nil {
		t.Errorf("SaveRates returned unexpected error: %v", err)
	}

	// GetLatestRates accepts context and currency string
	rates, err := repo.GetLatestRates(ctx, "USD")
	if err != nil {
		t.Errorf("GetLatestRates returned unexpected error: %v", err)
	}
	if rates != nil {
		t.Errorf("expected nil rates, got %v", rates)
	}

	// GetHistoryRates accepts context, currency, rateType, from, to, limit
	rates, err = repo.GetHistoryRates(ctx, "USD", "reference", "2026-01-01", "2026-07-10", 30)
	if err != nil {
		t.Errorf("GetHistoryRates returned unexpected error: %v", err)
	}
	if rates != nil {
		t.Errorf("expected nil rates, got %v", rates)
	}
}
