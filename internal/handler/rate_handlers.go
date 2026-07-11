package handler

import (
	"net/http"
	"strconv"
)

// GetRates handles GET /rates with optional currency and type query params.
func (h *Handler) GetRates(w http.ResponseWriter, r *http.Request) {
	currency := r.URL.Query().Get("currency")
	rateType := r.URL.Query().Get("type")

	rates, err := h.uc.GetCurrentRates(r.Context(), currency, rateType)
	if err != nil {
		status, code := mapError(err)
		respondError(w, status, code, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": rates})
}

// GetHistory handles GET /rates/history with filter query params.
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	currency := q.Get("currency")
	rateType := q.Get("type")
	from := q.Get("from")
	to := q.Get("to")

	limitStr := q.Get("limit")
	limit := 0
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid limit parameter")
			return
		}
	}

	rates, err := h.uc.GetHistoryRates(r.Context(), currency, rateType, from, to, limit)
	if err != nil {
		status, code := mapError(err)
		respondError(w, status, code, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": rates})
}

// TriggerScrape handles POST /admin/scrape.
func (h *Handler) TriggerScrape(w http.ResponseWriter, r *http.Request) {
	rates, err := h.uc.ScrapeRates(r.Context())
	if err != nil {
		status, code := mapError(err)
		respondError(w, status, code, "internal server error")
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"data": map[string]interface{}{
			"message":       "scrape triggered",
			"rates_scraped": len(rates),
		},
	})
}
