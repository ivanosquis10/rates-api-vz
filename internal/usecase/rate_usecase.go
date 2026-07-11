package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/scraper"
)

// RateUsecase orchestrates rate scraping and retrieval.
type RateUsecase struct {
	repo    domain.Repository
	scraper scraper.Scraper
}

// NewRateUsecase creates a RateUsecase with the given dependencies.
func NewRateUsecase(repo domain.Repository, s scraper.Scraper) *RateUsecase {
	return &RateUsecase{repo: repo, scraper: s}
}

// ScrapeRates calls the scraper and persists results.
func (uc *RateUsecase) ScrapeRates(ctx context.Context) ([]domain.Rate, error) {
	rates, err := uc.scraper.Scrape(ctx)
	if err != nil {
		slog.Error("scrape failed", "method", "ScrapeRates", "error", err)
		return nil, fmt.Errorf("ScrapeRates scrape: %w", err)
	}
	if err := uc.repo.SaveRates(ctx, rates); err != nil {
		slog.Error("save rates failed", "method", "ScrapeRates", "error", err)
		return nil, fmt.Errorf("ScrapeRates save: %w", err)
	}
	return rates, nil
}

// GetCurrentRates returns latest rates, optionally filtered by rateType.
func (uc *RateUsecase) GetCurrentRates(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
	rates, err := uc.repo.GetLatestRates(ctx, currency)
	if err != nil {
		slog.Error("get latest rates failed", "method", "GetCurrentRates", "error", err)
		return nil, fmt.Errorf("GetCurrentRates: %w", err)
	}
	if rateType == "" {
		return rates, nil
	}
	filtered := make([]domain.Rate, 0, len(rates))
	for _, r := range rates {
		if strings.EqualFold(r.RateType, rateType) {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// GetHistoryRates delegates to the repository with all filter parameters.
func (uc *RateUsecase) GetHistoryRates(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
	rates, err := uc.repo.GetHistoryRates(ctx, currency, rateType, from, to, limit)
	if err != nil {
		slog.Error("get history rates failed", "method", "GetHistoryRates", "error", err)
		return nil, fmt.Errorf("GetHistoryRates: %w", err)
	}
	return rates, nil
}
