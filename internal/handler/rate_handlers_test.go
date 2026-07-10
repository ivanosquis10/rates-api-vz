package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/middleware"
)

// mockUsecase is a minimal fake for testing handler delegation.
type mockUsecase struct {
	getCurrentRatesFn func(ctx context.Context, currency, rateType string) ([]domain.Rate, error)
	getHistoryRatesFn func(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error)
	scrapeRatesFn     func(ctx context.Context) (int, error)
}

func (m *mockUsecase) GetCurrentRates(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
	return m.getCurrentRatesFn(ctx, currency, rateType)
}

func (m *mockUsecase) GetHistoryRates(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
	return m.getHistoryRatesFn(ctx, currency, rateType, from, to, limit)
}

func (m *mockUsecase) ScrapeRates(ctx context.Context) (int, error) {
	return m.scrapeRatesFn(ctx)
}

// --- GetRates tests ---

func TestGetRates_NoFilter(t *testing.T) {
	mock := &mockUsecase{
		getCurrentRatesFn: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
			return []domain.Rate{
				{Currency: "USD", RateType: "reference", Value: 36.5},
				{Currency: "EUR", RateType: "buy", Value: 38.0},
			}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/rates", nil)
	w := httptest.NewRecorder()

	h.GetRates(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []domain.Rate `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 2 {
		t.Fatalf("expected 2 rates, got %d", len(body.Data))
	}
}

func TestGetRates_FilterByCurrency(t *testing.T) {
	mock := &mockUsecase{
		getCurrentRatesFn: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
			if currency != "USD" {
				t.Errorf("expected currency USD, got %s", currency)
			}
			return []domain.Rate{
				{Currency: "USD", RateType: "reference", Value: 36.5},
			}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/rates?currency=USD", nil)
	w := httptest.NewRecorder()

	h.GetRates(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []domain.Rate `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(body.Data))
	}
	if body.Data[0].Currency != "USD" {
		t.Errorf("expected USD, got %s", body.Data[0].Currency)
	}
}

func TestGetRates_FilterByCurrencyAndType(t *testing.T) {
	mock := &mockUsecase{
		getCurrentRatesFn: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
			if currency != "USD" || rateType != "reference" {
				t.Errorf("expected USD/reference, got %s/%s", currency, rateType)
			}
			return []domain.Rate{
				{Currency: "USD", RateType: "reference", Value: 36.5},
			}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/rates?currency=USD&type=reference", nil)
	w := httptest.NewRecorder()

	h.GetRates(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []domain.Rate `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(body.Data))
	}
}

// --- GetHistory tests ---

func TestGetHistory_WithAllFilters(t *testing.T) {
	mock := &mockUsecase{
		getHistoryRatesFn: func(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
			if currency != "USD" || rateType != "buy" || from != "2026-01-01" || to != "2026-07-01" || limit != 50 {
				t.Errorf("unexpected params: %s %s %s %s %d", currency, rateType, from, to, limit)
			}
			return []domain.Rate{
				{Currency: "USD", RateType: "buy", Value: 37.0},
			}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/rates/history?currency=USD&type=buy&from=2026-01-01&to=2026-07-01&limit=50", nil)
	w := httptest.NewRecorder()

	h.GetHistory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []domain.Rate `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(body.Data))
	}
}

func TestGetHistory_EmptyResult(t *testing.T) {
	mock := &mockUsecase{
		getHistoryRatesFn: func(ctx context.Context, currency, rateType, from, to string, limit int) ([]domain.Rate, error) {
			return []domain.Rate{}, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodGet, "/rates/history", nil)
	w := httptest.NewRecorder()

	h.GetHistory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data []domain.Rate `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(body.Data) != 0 {
		t.Fatalf("expected 0 rates, got %d", len(body.Data))
	}
}

func TestGetHistory_InvalidLimit(t *testing.T) {
	h := NewHandlerFromUsecaser(&mockUsecase{})
	req := httptest.NewRequest(http.MethodGet, "/rates/history?limit=abc", nil)
	w := httptest.NewRecorder()

	h.GetHistory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Error.Code != "BAD_REQUEST" {
		t.Errorf("expected BAD_REQUEST, got %s", body.Error.Code)
	}
}

// --- TriggerScrape tests ---

func TestTriggerScrape_Success(t *testing.T) {
	mock := &mockUsecase{
		scrapeRatesFn: func(ctx context.Context) (int, error) {
			return 12, nil
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodPost, "/admin/scrape", nil)
	w := httptest.NewRecorder()

	h.TriggerScrape(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}

	var body struct {
		Data struct {
			Message     string `json:"message"`
			RatesScraped int   `json:"rates_scraped"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Data.Message != "scrape triggered" {
		t.Errorf("expected 'scrape triggered', got %s", body.Data.Message)
	}
	if body.Data.RatesScraped != 12 {
		t.Errorf("expected 12 rates_scraped, got %d", body.Data.RatesScraped)
	}
}

func TestTriggerScrape_Error(t *testing.T) {
	mock := &mockUsecase{
		scrapeRatesFn: func(ctx context.Context) (int, error) {
			return 0, domain.ErrScrapeFailed
		},
	}

	h := NewHandlerFromUsecaser(mock)
	req := httptest.NewRequest(http.MethodPost, "/admin/scrape", nil)
	w := httptest.NewRecorder()

	h.TriggerScrape(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %s", body.Error.Code)
	}
}

// --- Phase 4: Integration verification tests (4.2-4.4) ---

// newTestRouter builds a Chi router with recovery middleware + all routes
// wired to a mock usecase. Used by integration verification tests.
func newTestRouter(mock Usecaser) *chi.Mux {
	h := NewHandlerFromUsecaser(mock)
	r := chi.NewRouter()
	r.Use(middleware.Recovery)
	r.Get("/rates", h.GetRates)
	r.Get("/rates/history", h.GetHistory)
	r.Route("/admin", func(r chi.Router) {
		r.Post("/scrape", h.TriggerScrape)
	})
	return r
}

// 4.2 — Panic recovery returns 500 without crashing server.
func TestVerification_PanicRecoveryReturns500(t *testing.T) {
	// Handler that panics — recovery middleware must catch it.
	mock := &mockUsecase{
		getCurrentRatesFn: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
			panic("unexpected nil pointer")
		},
	}

	router := newTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/rates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after panic, got %d", w.Code)
	}

	// Verify server is still alive by sending a second request.
	req2 := httptest.NewRequest(http.MethodGet, "/rates", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusInternalServerError {
		t.Fatalf("expected server to survive panic and return 500, got %d", w2.Code)
	}
}

// 4.3 — 500 responses never contain internal error details.
func TestVerification_500ResponsesSanitized(t *testing.T) {
	internalErr := errors.New("pq: relation \"rates\" does not exist")
	mock := &mockUsecase{
		getCurrentRatesFn: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
			return nil, internalErr
		},
	}

	router := newTestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/rates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	body := w.Body.String()
	// Must NOT leak internal details.
	if strings.Contains(body, "pq:") {
		t.Errorf("response leaks SQL error: %s", body)
	}
	if strings.Contains(body, "relation") {
		t.Errorf("response leaks database detail: %s", body)
	}
	// Must contain sanitized message.
	if !strings.Contains(body, "internal server error") {
		t.Errorf("expected generic message, got: %s", body)
	}
}

// 4.4 — All responses use { "data": ... } or { "error": { "code", "message" } } envelope.
func TestVerification_ResponseEnvelopeConsistency(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		mock    *mockUsecase
		wantKey string // "data" or "error"
	}{
		{
			name:   "success uses data envelope",
			method: http.MethodGet,
			path:   "/rates",
			mock: &mockUsecase{
				getCurrentRatesFn: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
					return []domain.Rate{{Currency: "USD", RateType: "reference", Value: 36.5}}, nil
				},
			},
			wantKey: "data",
		},
		{
			name:   "error uses error envelope",
			method: http.MethodGet,
			path:   "/rates",
			mock: &mockUsecase{
				getCurrentRatesFn: func(ctx context.Context, currency, rateType string) ([]domain.Rate, error) {
					return nil, domain.ErrNotFound
				},
			},
			wantKey: "error",
		},
		{
			name:   "400 uses error envelope",
			method: http.MethodGet,
			path:   "/rates/history?limit=abc",
			mock:   &mockUsecase{},
			wantKey: "error",
		},
		{
			name:   "202 uses data envelope",
			method: http.MethodPost,
			path:   "/admin/scrape",
			mock: &mockUsecase{
				scrapeRatesFn: func(ctx context.Context) (int, error) {
					return 5, nil
				},
			},
			wantKey: "data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(tt.mock)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var raw map[string]json.RawMessage
			if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if _, ok := raw[tt.wantKey]; !ok {
				t.Errorf("expected key %q in response, got keys: %v", tt.wantKey, keysOf(raw))
			}
			// Ensure the OTHER key is absent.
			other := "error"
			if tt.wantKey == "error" {
				other = "data"
			}
			if _, ok := raw[other]; ok {
				t.Errorf("unexpected key %q in response with key %q", other, tt.wantKey)
			}
		})
	}
}

func keysOf(m map[string]json.RawMessage) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
