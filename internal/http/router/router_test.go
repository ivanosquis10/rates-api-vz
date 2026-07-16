package router_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ivanosquis10/api-rates-venezuela/internal/config"
	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/handler"
	"github.com/ivanosquis10/api-rates-venezuela/internal/http/router"
	"github.com/ivanosquis10/api-rates-venezuela/internal/middleware"
)

type mockUsecase struct {
	getLatestRateFn   func(ctx context.Context, currency string) (domain.Rate, error)
	getHistoryRatesFn func(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error)
	scrapeRatesFn     func(ctx context.Context) ([]domain.Rate, error)
}

func (m *mockUsecase) GetLatestRate(ctx context.Context, currency string) (domain.Rate, error) {
	if m.getLatestRateFn != nil {
		return m.getLatestRateFn(ctx, currency)
	}
	return domain.Rate{}, nil
}

func (m *mockUsecase) GetHistoryRates(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error) {
	if m.getHistoryRatesFn != nil {
		return m.getHistoryRatesFn(ctx, currency, from, to, limit)
	}
	return nil, nil
}

func (m *mockUsecase) ScrapeRates(ctx context.Context) ([]domain.Rate, error) {
	if m.scrapeRatesFn != nil {
		return m.scrapeRatesFn(ctx)
	}
	return nil, nil
}

func TestRouter_New(t *testing.T) {
	mockUC := &mockUsecase{}
	h := handler.NewHandlerFromUsecaser(mockUC)
	cfg := &config.Config{
		APIKey:                "test-api-key",
		RateLimit:             100,
		ScrapeCronMaintenance: "0 8 * * *",
		ScrapeCronWindow:      "*/5 8-18 * * 1-5",
		Port:                  8080,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rl := middleware.NewRateLimiter(ctx, cfg.RateLimit)

	deps := router.Deps{
		Handler:     h,
		Config:      cfg,
		RateLimiter: rl,
	}

	r := router.New(deps)
	if r == nil {
		t.Fatal("expected router to be non-nil")
	}
}

func TestRouter_Middleware_Auth(t *testing.T) {
	mockUC := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			return domain.Rate{Currency: "USD", Value: 36.5}, nil
		},
		getHistoryRatesFn: func(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error) {
			return []domain.Rate{}, nil
		},
		scrapeRatesFn: func(ctx context.Context) ([]domain.Rate, error) {
			return []domain.Rate{}, nil
		},
	}
	h := handler.NewHandlerFromUsecaser(mockUC)
	cfg := &config.Config{
		APIKey:                "secret-api-key",
		RateLimit:             100,
		ScrapeCronMaintenance: "0 8 * * *",
		ScrapeCronWindow:      "*/5 8-18 * * 1-5",
		Port:                  8080,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rl := middleware.NewRateLimiter(ctx, cfg.RateLimit)

	deps := router.Deps{
		Handler:     h,
		Config:      cfg,
		RateLimiter: rl,
	}

	r := router.New(deps)

	tests := []struct {
		name           string
		method         string
		url            string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "Rates request without API Key",
			method:         "GET",
			url:            "/api/v1/dollars",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Rates request with invalid API Key",
			method:         "GET",
			url:            "/api/v1/dollars",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Rates request with valid API Key",
			method:         "GET",
			url:            "/api/v1/dollars",
			apiKey:         "secret-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "History request with valid API Key",
			method:         "GET",
			url:            "/api/v1/history/dollars",
			apiKey:         "secret-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Scrape request with valid API Key",
			method:         "POST",
			url:            "/api/v1/admin/scrape",
			apiKey:         "secret-api-key",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Header().Get("X-Request-ID") == "" {
				t.Error("expected X-Request-ID response header to be set, but it was empty")
			}
		})
	}
}

func TestRouter_NotFound(t *testing.T) {
	mockUC := &mockUsecase{}
	h := handler.NewHandlerFromUsecaser(mockUC)
	cfg := &config.Config{
		APIKey:                "secret-api-key",
		RateLimit:             100,
		ScrapeCronMaintenance: "0 8 * * *",
		ScrapeCronWindow:      "*/5 8-18 * * 1-5",
		Port:                  8080,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rl := middleware.NewRateLimiter(ctx, cfg.RateLimit)

	deps := router.Deps{
		Handler:     h,
		Config:      cfg,
		RateLimiter: rl,
	}

	r := router.New(deps)

	req, err := http.NewRequest(http.MethodGet, "/invalid-route", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-API-Key", "secret-api-key")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if raw["success"] != false {
		t.Errorf("expected success to be false, got %v", raw["success"])
	}
	if raw["code"] != "NOT_FOUND" {
		t.Errorf("expected code to be NOT_FOUND, got %v", raw["code"])
	}
}
