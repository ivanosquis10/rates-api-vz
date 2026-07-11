package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth_Success(t *testing.T) {
	h := NewHandlerFromUsecaser(&mockUsecase{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Status  string `json:"status"`
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !body.Success {
		t.Error("expected success to be true")
	}
	if body.Data.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", body.Data.Status)
	}
	if body.Data.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", body.Data.Version)
	}
}

func TestHealth_ContentType(t *testing.T) {
	h := NewHandlerFromUsecaser(&mockUsecase{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestHealth_EnvelopeStructure(t *testing.T) {
	h := NewHandlerFromUsecaser(&mockUsecase{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Must have success and data
	if _, exists := raw["success"]; !exists {
		t.Error("expected 'success' key in response")
	}
	if _, exists := raw["data"]; !exists {
		t.Error("expected 'data' key in response")
	}
	// Must NOT have error or code keys
	if _, exists := raw["error"]; exists {
		t.Error("expected 'error' key to be absent on success")
	}
	if _, exists := raw["code"]; exists {
		t.Error("expected 'code' key to be absent on success")
	}
}
