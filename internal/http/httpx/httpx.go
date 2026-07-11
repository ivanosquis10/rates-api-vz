package httpx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// WriteJSON sets Content-Type to application/json, writes the status code, and serializes data.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// GetRequestID wraps Chi's middleware helper to extract request ID from context.
func GetRequestID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}
