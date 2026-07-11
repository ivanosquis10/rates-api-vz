package router_test

import (
	"context"
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
	getCurrentRatesFunc func(ctx context.Context, currency, rateType string) ([]domain.Rate, error)
	getHistoryRatesFunc func(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error)
	scrapeRatesFunc     func(ctx context.Context) ([]domain.Rate, error)
}

func (m *mockUsecase) GetCurrentRates(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
	if m.getCurrentRatesFunc != nil {
		return m.getCurrentRatesFunc(ctx, currency, rateType)
	}
	return nil, nil
}

func (m *mockUsecase) GetHistoryRates(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
	if m.getHistoryRatesFunc != nil {
		return m.getHistoryRatesFunc(ctx, currency, rateType, from, to, limit)
	}
	return nil, nil
}

func (m *mockUsecase) ScrapeRates(ctx context.Context) ([]domain.Rate, error) {
	if m.scrapeRatesFunc != nil {
		return m.scrapeRatesFunc(ctx)
	}
	return nil, nil
}

func TestRouter_New(t *testing.T) {
	mockUC := &mockUsecase{}
	h := handler.NewHandlerFromUsecaser(mockUC)
	cfg := &config.Config{
		APIKey:     "test-api-key",
		RateLimit:  100,
		ScrapeHour: 9,
		Port:       8080,
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
		getCurrentRatesFunc: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
			return []domain.Rate{}, nil
		},
		getHistoryRatesFunc: func(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
			return []domain.Rate{}, nil
		},
		scrapeRatesFunc: func(ctx context.Context) ([]domain.Rate, error) {
			return []domain.Rate{}, nil
		},
	}
	h := handler.NewHandlerFromUsecaser(mockUC)
	cfg := &config.Config{
		APIKey:     "secret-api-key",
		RateLimit:  100,
		ScrapeHour: 9,
		Port:       8080,
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
			url:            "/rates",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Rates request with invalid API Key",
			method:         "GET",
			url:            "/rates",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Rates request with valid API Key",
			method:         "GET",
			url:            "/rates",
			apiKey:         "secret-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "History request with valid API Key",
			method:         "GET",
			url:            "/rates/history",
			apiKey:         "secret-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Scrape request with valid API Key",
			method:         "POST",
			url:            "/admin/scrape",
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
		})
	}
}
