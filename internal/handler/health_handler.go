package handler

import (
	"net/http"

	"github.com/ivanosquis10/api-rates-venezuela/internal/presenter"
)

const apiVersion = "1.0.0"

// HealthResponse is the payload returned by the health endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Health returns a lightweight health check response.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	presenter.OK(w, r, HealthResponse{
		Status:  "ok",
		Version: apiVersion,
	})
}
