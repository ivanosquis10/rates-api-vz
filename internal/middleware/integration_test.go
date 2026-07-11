package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func TestMiddleware_ExecutionOrder(t *testing.T) {
	// Order: Recovery -> Logging -> RateLimit -> Auth.
	// We verify execution order by asserting status code precedence:
	// - RateLimit is 3rd, Auth is 4th.
	// - If RateLimit is before Auth, an request from an IP that has exceeded
	//   its limit will trigger 429 even if it has an invalid/missing API key.
	// - A request from a fresh IP will pass RateLimit and fail Auth with 401.

	type responseEnvelope struct {
		Success bool    `json:"success"`
		Data    any     `json:"data"`
		Code    *string `json:"code"`
		Error   *string `json:"error"`
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiKey := "integration-test-key"
	rateLimitMw := NewRateLimiter(ctx, 1).Handler // 1 req/min limit, burst = 1
	authMw := Auth(apiKey)

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Wrap in sequence: requestID -> recovery -> logging -> rateLimit -> auth
	handler := chimw.RequestID(Recovery(Logging(rateLimitMw(authMw(finalHandler)))))

	// Scenario A: Valid request (passes all middlewares)
	{
		req := httptest.NewRequest(http.MethodGet, "/rates", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	}

	// Scenario B: Rate limit exceeded (2nd request from same IP)
	// It should return 429 because RateLimit is executed BEFORE Auth (precedence check).
	{
		req := httptest.NewRequest(http.MethodGet, "/rates", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		// Not setting X-API-Key
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected 429 (rate limit before auth check), got %d", w.Code)
		}

		if w.Header().Get("X-Request-ID") == "" {
			t.Error("expected X-Request-ID response header to be set, but it was empty")
		}

		var resp responseEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}
		if resp.Success {
			t.Error("expected success to be false")
		}
		if resp.Code == nil || *resp.Code != "RATE_LIMITED" {
			t.Errorf("expected code RATE_LIMITED, got %v", resp.Code)
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

	// Scenario C: Auth failure for a new IP (not rate limited)
	// A new IP has not exceeded rate limit, so RateLimit passes, and Auth fails, returning 401.
	{
		req := httptest.NewRequest(http.MethodGet, "/rates", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		// Not setting X-API-Key
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 (auth failure after rate limit pass), got %d", w.Code)
		}

		if w.Header().Get("X-Request-ID") == "" {
			t.Error("expected X-Request-ID response header to be set, but it was empty")
		}

		var resp responseEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}
		if resp.Success {
			t.Error("expected success to be false")
		}
		if resp.Code == nil || *resp.Code != "UNAUTHORIZED" {
			t.Errorf("expected code UNAUTHORIZED, got %v", resp.Code)
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
}
