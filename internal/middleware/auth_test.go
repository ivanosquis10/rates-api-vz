package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type responseEnvelope struct {
	Success bool    `json:"success"`
	Data    any     `json:"data"`
	Code    *string `json:"code"`
	Error   *string `json:"error"`
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
	ctx := context.WithValue(req.Context(), chimw.RequestIDKey, "test-auth-req-id")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
	if called {
		t.Error("expected next handler NOT to be called")
	}

	if got := w.Header().Get("X-Request-ID"); got != "test-auth-req-id" {
		t.Errorf("expected X-Request-ID header %q, got %q", "test-auth-req-id", got)
	}

	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success to be false, got true")
	}
	if resp.Code == nil || *resp.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %v", resp.Code)
	}
	if resp.Error == nil || *resp.Error != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %v", resp.Error)
	}

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal raw response: %v", err)
	}
	if _, exists := raw["data"]; exists {
		t.Error("expected 'data' key to be omitted on error")
	}
	if _, exists := raw["request_id"]; exists {
		t.Error("expected 'request_id' key to be omitted from JSON body")
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
	ctx := context.WithValue(req.Context(), chimw.RequestIDKey, "test-auth-req-id")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
	if called {
		t.Error("expected next handler NOT to be called")
	}

	if got := w.Header().Get("X-Request-ID"); got != "test-auth-req-id" {
		t.Errorf("expected X-Request-ID header %q, got %q", "test-auth-req-id", got)
	}

	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success to be false, got true")
	}
	if resp.Code == nil || *resp.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %v", resp.Code)
	}
	if resp.Error == nil || *resp.Error != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %v", resp.Error)
	}

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal raw response: %v", err)
	}
	if _, exists := raw["data"]; exists {
		t.Error("expected 'data' key to be omitted on error")
	}
	if _, exists := raw["request_id"]; exists {
		t.Error("expected 'request_id' key to be omitted from JSON body")
	}
}
