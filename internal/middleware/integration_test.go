package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_ExecutionOrder(t *testing.T) {
	// Order: Recovery -> Logging -> RateLimit -> Auth.
	// We verify execution order by asserting status code precedence:
	// - RateLimit is 3rd, Auth is 4th.
	// - If RateLimit is before Auth, an request from an IP that has exceeded
	//   its limit will trigger 429 even if it has an invalid/missing API key.
	// - A request from a fresh IP will pass RateLimit and fail Auth with 401.

	type responseEnvelope struct {
		Success   bool    `json:"success"`
		Data      any     `json:"data"`
		ErrorCode *string `json:"error_code"`
		Error     *string `json:"error"`
		RequestID string  `json:"request_id"`
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

	// Wrap in sequence: recovery -> logging -> rateLimit -> auth
	handler := Recovery(Logging(rateLimitMw(authMw(finalHandler))))

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

		var resp responseEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}
		if resp.Success {
			t.Error("expected success to be false")
		}
		if resp.ErrorCode == nil || *resp.ErrorCode != "RATE_LIMITED" {
			t.Errorf("expected error_code RATE_LIMITED, got %v", resp.ErrorCode)
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

		var resp responseEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}
		if resp.Success {
			t.Error("expected success to be false")
		}
		if resp.ErrorCode == nil || *resp.ErrorCode != "UNAUTHORIZED" {
			t.Errorf("expected error_code UNAUTHORIZED, got %v", resp.ErrorCode)
		}
	}
}
