package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// respondError writes a standardized error envelope.
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

// mapError converts a domain error to HTTP status and error code.
func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, "BAD_REQUEST"
	case errors.Is(err, domain.ErrDuplicateRate):
		return http.StatusConflict, "CONFLICT"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
