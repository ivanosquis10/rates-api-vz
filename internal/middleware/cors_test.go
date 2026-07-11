package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_GET(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(inner)

	req := httptest.NewRequest(http.MethodGet, "/rates", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods: GET, POST, OPTIONS, got: %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, X-API-Key" {
		t.Errorf("expected Access-Control-Allow-Headers: Content-Type, X-API-Key, got: %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORS_Preflight(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(inner)

	req := httptest.NewRequest(http.MethodOptions, "/rates", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods: GET, POST, OPTIONS, got: %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, X-API-Key" {
		t.Errorf("expected Access-Control-Allow-Headers: Content-Type, X-API-Key, got: %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORS_POST(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(inner)

	req := httptest.NewRequest(http.MethodPost, "/admin/scrape", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods: GET, POST, OPTIONS, got: %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, X-API-Key" {
		t.Errorf("expected Access-Control-Allow-Headers: Content-Type, X-API-Key, got: %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
}
