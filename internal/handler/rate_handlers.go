package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ivanosquis10/api-rates-venezuela/internal/apierrors"
	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/presenter"
)

// GetUSD handles GET /dollars. Returns a JSON array containing the latest USD rate.
func (h *Handler) GetUSD(w http.ResponseWriter, r *http.Request) {
	rate, err := h.uc.GetLatestRate(r.Context(), "USD")
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			presenter.OK(w, r, []presenter.RateResponse{})
			return
		}
		presenter.Error(w, r, err)
		return
	}

	presenter.OK(w, r, []presenter.RateResponse{presenter.MapToRateResponse(rate)})
}

// GetOfficialUSD handles GET /dollars/official. Returns a single JSON object of the latest USD rate.
func (h *Handler) GetOfficialUSD(w http.ResponseWriter, r *http.Request) {
	rate, err := h.uc.GetLatestRate(r.Context(), "USD")
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.OK(w, r, presenter.MapToRateResponse(rate))
}

// GetEUR handles GET /euros. Returns a JSON array containing the latest EUR rate.
func (h *Handler) GetEUR(w http.ResponseWriter, r *http.Request) {
	rate, err := h.uc.GetLatestRate(r.Context(), "EUR")
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			presenter.OK(w, r, []presenter.RateResponse{})
			return
		}
		presenter.Error(w, r, err)
		return
	}

	presenter.OK(w, r, []presenter.RateResponse{presenter.MapToRateResponse(rate)})
}

// GetOfficialEUR handles GET /euros/official. Returns a single JSON object of the latest EUR rate.
func (h *Handler) GetOfficialEUR(w http.ResponseWriter, r *http.Request) {
	rate, err := h.uc.GetLatestRate(r.Context(), "EUR")
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.OK(w, r, presenter.MapToRateResponse(rate))
}

// GetUSDHistory handles GET /history/dollars. Returns historical USD rates.
func (h *Handler) GetUSDHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")

	limitStr := q.Get("limit")
	limit := 0
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			presenter.Error(w, r, apierrors.New(apierrors.BAD_REQUEST, "invalid limit parameter"))
			return
		}
	}

	rates, err := h.uc.GetHistoryRates(r.Context(), "USD", from, to, limit)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.OK(w, r, presenter.MapToRateResponses(rates))
}

// GetEURHistory handles GET /history/euros. Returns historical EUR rates.
func (h *Handler) GetEURHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")

	limitStr := q.Get("limit")
	limit := 0
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			presenter.Error(w, r, apierrors.New(apierrors.BAD_REQUEST, "invalid limit parameter"))
			return
		}
	}

	rates, err := h.uc.GetHistoryRates(r.Context(), "EUR", from, to, limit)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.OK(w, r, presenter.MapToRateResponses(rates))
}

// TriggerScrape handles POST /admin/scrape.
func (h *Handler) TriggerScrape(w http.ResponseWriter, r *http.Request) {
	rates, err := h.uc.ScrapeRates(r.Context())
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.OK(w, r, map[string]any{
		"message":       "scrape triggered",
		"rates_scraped": len(rates),
	})
}
