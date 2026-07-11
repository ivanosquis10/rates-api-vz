package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimit_LimitCapacity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2 req/min: rate = 2/60 = 1/30 tokens/sec. Burst = 2.
	rl := NewRateLimiter(ctx, 2)

	calledCount := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledCount++
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Handler(next)

	// 1st request
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("request 1: expected 200, got %d", w1.Code)
	}

	// 2nd request
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("request 2: expected 200, got %d", w2.Code)
	}

	// 3rd request (should exceed limit)
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "1.2.3.4:1234"
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("request 3: expected 429, got %d", w3.Code)
	}

	retryAfterStr := w3.Header().Get("Retry-After")
	if retryAfterStr == "" {
		t.Error("expected Retry-After header to be present")
	} else {
		retryAfter, err := strconv.Atoi(retryAfterStr)
		if err != nil {
			t.Errorf("expected Retry-After to be an integer, got %s: %v", retryAfterStr, err)
		}
		if retryAfter <= 0 {
			t.Errorf("expected Retry-After to be > 0, got %d", retryAfter)
		}
	}

	var errorResp responseEnvelope
	if err := json.Unmarshal(w3.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if errorResp.Success {
		t.Error("expected success to be false")
	}
	if errorResp.ErrorCode == nil || *errorResp.ErrorCode != "RATE_LIMITED" {
		t.Errorf("expected error_code RATE_LIMITED, got %v", errorResp.ErrorCode)
	}
	if errorResp.Error == nil || *errorResp.Error != "too many requests" {
		t.Errorf("expected error message 'too many requests', got %v", errorResp.Error)
	}
}

func TestRateLimit_SeparateIPs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rl := NewRateLimiter(ctx, 1)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Handler(next)

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "1.1.1.1:1234"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("IP 1 request 1: expected 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "1.1.1.1:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("IP 1 request 2: expected 429, got %d", w2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "2.2.2.2:1234"
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("IP 2 request 1: expected 200, got %d", w3.Code)
	}
}

func TestRateLimit_InvalidIPFallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rl := NewRateLimiter(ctx, 1)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Handler(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "invalid_ip_format"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRateLimit_JanitorLifecycle(t *testing.T) {
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	startGoroutines := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = NewRateLimiter(ctx, 10)

	time.Sleep(50 * time.Millisecond)

	endGoroutines := runtime.NumGoroutine()
	t.Logf("startGoroutines: %d, endGoroutines: %d", startGoroutines, endGoroutines)
	if endGoroutines <= startGoroutines {
		t.Error("expected janitor goroutine to start")
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestRateLimit_Pruning(t *testing.T) {
	rl := &RateLimiter{
		clients:     make(map[string]*client),
		limitPerMin: 60,
		pruneAge:    1 * time.Second,
	}

	c1 := &client{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: time.Now().Unix() - 5,
	}
	c2 := &client{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: time.Now().Unix(),
	}

	rl.clients["1.1.1.1"] = c1
	rl.clients["2.2.2.2"] = c2

	rl.prune()

	if _, exists := rl.clients["1.1.1.1"]; exists {
		t.Error("expected 1.1.1.1 to be pruned")
	}
	if _, exists := rl.clients["2.2.2.2"]; !exists {
		t.Error("expected 2.2.2.2 NOT to be pruned")
	}
}

func TestRateLimit_Races(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rl := NewRateLimiter(ctx, 1000)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Handler(next)

	var wg sync.WaitGroup
	const numGoroutines = 10
	const numRequests = 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = fmt.Sprintf("192.168.1.%d:1234", id)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
			}
		}(i)
	}
	wg.Wait()
}
