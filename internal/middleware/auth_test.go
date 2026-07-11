package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type responseEnvelope struct {
	Success   bool    `json:"success"`
	Data      any     `json:"data"`
	ErrorCode *string `json:"error_code"`
	Error     *string `json:"error"`
	RequestID string  `json:"request_id"`
}

func TestAuth_ValidKey(t *testing.T) {
	apiKey := "secret-test-key"
	authMw := Auth(apiKey)

	called := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := authMw(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/rates", nil)
	req.Header.Set("X-API-Key", apiKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Error("expected next handler to be called")
	}
}

func TestAuth_MissingKey(t *testing.T) {
	apiKey := "secret-test-key"
	authMw := Auth(apiKey)

	called := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := authMw(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/rates", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
	if called {
		t.Error("expected next handler NOT to be called")
	}

	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success to be false, got true")
	}
	if resp.ErrorCode == nil || *resp.ErrorCode != "UNAUTHORIZED" {
		t.Errorf("expected error_code UNAUTHORIZED, got %v", resp.ErrorCode)
	}
	if resp.Error == nil || *resp.Error != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %v", resp.Error)
	}
}

func TestAuth_InvalidKey(t *testing.T) {
	apiKey := "secret-test-key"
	authMw := Auth(apiKey)

	called := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := authMw(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/rates", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
	if called {
		t.Error("expected next handler NOT to be called")
	}

	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success to be false, got true")
	}
	if resp.ErrorCode == nil || *resp.ErrorCode != "UNAUTHORIZED" {
		t.Errorf("expected error_code UNAUTHORIZED, got %v", resp.ErrorCode)
	}
	if resp.Error == nil || *resp.Error != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %v", resp.Error)
	}
}
