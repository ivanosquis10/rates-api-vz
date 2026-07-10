package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

// --- respondJSON tests ---

func TestRespondJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()

	respondJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", body["key"])
	}
}

func TestRespondJSON_DataEnvelope(t *testing.T) {
	w := httptest.NewRecorder()

	rates := []domain.Rate{
		{Currency: "USD", RateType: "reference", Value: 36.5},
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": rates})

	var body struct {
		Data []domain.Rate `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Data) != 1 {
		t.Errorf("expected 1 rate in data, got %d", len(body.Data))
	}
	if body.Data[0].Currency != "USD" {
		t.Errorf("expected currency USD, got %s", body.Data[0].Currency)
	}
}

// --- respondError tests ---

func TestRespondError_Envelope(t *testing.T) {
	w := httptest.NewRecorder()

	respondError(w, http.StatusNotFound, "NOT_FOUND", "rate not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code NOT_FOUND, got %s", body.Error.Code)
	}
	if body.Error.Message != "rate not found" {
		t.Errorf("expected message 'rate not found', got %s", body.Error.Message)
	}
}

// --- mapError tests ---

func TestMapError_NotFound(t *testing.T) {
	status, code := mapError(domain.ErrNotFound)
	if status != http.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
	if code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", code)
	}
}

func TestMapError_InvalidInput(t *testing.T) {
	status, code := mapError(domain.ErrInvalidInput)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
	if code != "BAD_REQUEST" {
		t.Errorf("expected BAD_REQUEST, got %s", code)
	}
}

func TestMapError_DuplicateRate(t *testing.T) {
	status, code := mapError(domain.ErrDuplicateRate)
	if status != http.StatusConflict {
		t.Errorf("expected 409, got %d", status)
	}
	if code != "CONFLICT" {
		t.Errorf("expected CONFLICT, got %s", code)
	}
}

func TestMapError_Unknown(t *testing.T) {
	err := errors.New("something unexpected")
	status, code := mapError(err)
	if status != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
	if code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %s", code)
	}
}
