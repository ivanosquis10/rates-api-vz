package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/middleware"
	"github.com/ivanosquis10/api-rates-venezuela/internal/presenter"
)

// mockUsecase is a minimal fake for testing handler delegation.
type mockUsecase struct {
	getLatestRateFn   func(ctx context.Context, currency string) (domain.Rate, error)
	getHistoryRatesFn func(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error)
	scrapeRatesFn     func(ctx context.Context) ([]domain.Rate, error)
}

func (m *mockUsecase) GetLatestRate(ctx context.Context, currency string) (domain.Rate, error) {
	return m.getLatestRateFn(ctx, currency)
}

func (m *mockUsecase) GetHistoryRates(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error) {
	return m.getHistoryRatesFn(ctx, currency, from, to, limit)
}

func (m *mockUsecase) ScrapeRates(ctx context.Context) ([]domain.Rate, error) {
	return m.scrapeRatesFn(ctx)
}

func TestGetUSD(t *testing.T) {
	now := time.Now()
	mock := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			if currency != "USD" {
				t.Errorf("expected USD, got %s", currency)
			}
			return domain.Rate{ID: 1, Currency: "USD", Value: 36.5, ScrapedAt: now}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/dollars", nil)
	w := httptest.NewRecorder()

	h.GetUSD(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []presenter.RateResponse `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 rate in array, got %d", len(body.Data))
	}
	if body.Data[0].Currency != "USD" {
		t.Errorf("expected USD, got %s", body.Data[0].Currency)
	}
	if body.Data[0].Average != 36.5 {
		t.Errorf("expected average=36.5, got %f", body.Data[0].Average)
	}
}

func TestGetOfficialUSD(t *testing.T) {
	now := time.Now()
	mock := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			return domain.Rate{ID: 2, Currency: "USD", Value: 36.5, ScrapedAt: now}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/dollars/official", nil)
	w := httptest.NewRecorder()

	h.GetOfficialUSD(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data presenter.RateResponse `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Data.Currency != "USD" {
		t.Errorf("expected USD, got %s", body.Data.Currency)
	}
	if body.Data.Average != 36.5 {
		t.Errorf("expected average=36.5, got %f", body.Data.Average)
	}
}

func TestGetEUR(t *testing.T) {
	now := time.Now()
	mock := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			if currency != "EUR" {
				t.Errorf("expected EUR, got %s", currency)
			}
			return domain.Rate{ID: 3, Currency: "EUR", Value: 40.2, ScrapedAt: now}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/euros", nil)
	w := httptest.NewRecorder()

	h.GetEUR(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []presenter.RateResponse `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 rate in array, got %d", len(body.Data))
	}
	if body.Data[0].Currency != "EUR" {
		t.Errorf("expected EUR, got %s", body.Data[0].Currency)
	}
	if body.Data[0].Average != 40.2 {
		t.Errorf("expected average=40.2, got %f", body.Data[0].Average)
	}
}

func TestGetOfficialEUR(t *testing.T) {
	now := time.Now()
	mock := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			return domain.Rate{ID: 4, Currency: "EUR", Value: 40.2, ScrapedAt: now}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/euros/official", nil)
	w := httptest.NewRecorder()

	h.GetOfficialEUR(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data presenter.RateResponse `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Data.Currency != "EUR" {
		t.Errorf("expected EUR, got %s", body.Data.Currency)
	}
	if body.Data.Average != 40.2 {
		t.Errorf("expected average=40.2, got %f", body.Data.Average)
	}
}

func TestGetUSDHistory_Success(t *testing.T) {
	mock := &mockUsecase{
		getHistoryRatesFn: func(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error) {
			if currency != "USD" || from != "2026-01-01" || to != "2026-07-01" || limit != 50 {
				t.Errorf("unexpected params: %s %s %s %d", currency, from, to, limit)
			}
			return []domain.Rate{
				{ID: 5, Currency: "USD", Value: 37.0},
			}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/history/dollars?from=2026-01-01&to=2026-07-01&limit=50", nil)
	w := httptest.NewRecorder()

	h.GetUSDHistory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []presenter.RateResponse `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(body.Data))
	}
	if body.Data[0].Average != 37.0 {
		t.Errorf("expected average=37.0, got %f", body.Data[0].Average)
	}
}

func TestGetUSDHistory_InvalidLimit(t *testing.T) {
	h := NewHandlerFromUsecaser(&mockUsecase{})
	req := httptest.NewRequest(http.MethodGet, "/history/dollars?limit=abc", nil)
	w := httptest.NewRecorder()

	h.GetUSDHistory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTriggerScrape_Success(t *testing.T) {
	mock := &mockUsecase{
		scrapeRatesFn: func(ctx context.Context) ([]domain.Rate, error) {
			return make([]domain.Rate, 2), nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodPost, "/admin/scrape", nil)
	w := httptest.NewRecorder()

	h.TriggerScrape(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Message      string `json:"message"`
			RatesScraped int    `json:"rates_scraped"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !body.Success {
		t.Error("expected success to be true")
	}
	if body.Data.RatesScraped != 2 {
		t.Errorf("expected 2 rates_scraped, got %d", body.Data.RatesScraped)
	}
}

func TestTriggerScrape_Error(t *testing.T) {
	mock := &mockUsecase{
		scrapeRatesFn: func(ctx context.Context) ([]domain.Rate, error) {
			return nil, domain.ErrScrapeFailed
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodPost, "/admin/scrape", nil)
	w := httptest.NewRecorder()

	h.TriggerScrape(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func newTestRouter(mock Usecaser) *chi.Mux {
	h := NewHandlerFromUsecaser(mock)
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(middleware.Recovery)
	r.Get("/dollars", h.GetUSD)
	r.Get("/dollars/official", h.GetOfficialUSD)
	r.Route("/admin", func(r chi.Router) {
		r.Post("/scrape", h.TriggerScrape)
	})
	return r
}

func TestVerification_PanicRecoveryReturns500(t *testing.T) {
	mock := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			panic("unexpected nil pointer")
		},
	}

	router := newTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/dollars", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after panic, got %d", w.Code)
	}
}

func TestVerification_500ResponsesSanitized(t *testing.T) {
	internalErr := errors.New("pq: relation \"rates\" does not exist")
	mock := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			return domain.Rate{}, internalErr
		},
	}

	router := newTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/dollars", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "pq:") {
		t.Errorf("response leaks SQL error: %s", body)
	}
	if !strings.Contains(body, "internal server error") {
		t.Errorf("expected generic message, got: %s", body)
	}
}

func TestVerification_ResponseEnvelopeConsistency(t *testing.T) {
	type verificationEnvelope struct {
		Success bool    `json:"success"`
		Data    any     `json:"data"`
		Code    *string `json:"code"`
		Error   *string `json:"error"`
	}

	mock := &mockUsecase{
		getLatestRateFn: func(ctx context.Context, currency string) (domain.Rate, error) {
			if currency == "EUR" {
				return domain.Rate{}, domain.ErrNotFound
			}
			return domain.Rate{Currency: "USD", Value: 36.5}, nil
		},
	}

	router := chi.NewRouter()
	h := NewHandlerFromUsecaser(mock)
	router.Use(chimw.RequestID)
	router.Use(middleware.Recovery)
	router.Get("/dollars", h.GetUSD)
	router.Get("/euros", h.GetEUR)

	// Test Success
	req1 := httptest.NewRequest(http.MethodGet, "/dollars", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var env1 verificationEnvelope
	if err := json.Unmarshal(w1.Body.Bytes(), &env1); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !env1.Success {
		t.Error("expected success to be true")
	}

	// Test Error
	req2 := httptest.NewRequest(http.MethodGet, "/euros", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var env2 verificationEnvelope
	if err := json.Unmarshal(w2.Body.Bytes(), &env2); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env2.Success {
		t.Error("expected success to be false")
	}
}
