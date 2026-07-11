package presenter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/ivanosquis10/api-rates-venezuela/internal/apierrors"
)

func TestErrorPresenter(t *testing.T) {
	tests := []struct {
		name           string
		inputErr       error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "Standard Provider Error defaults to 502",
			inputErr:       apierrors.NewProviderError(errors.New("something went wrong on BCV")),
			expectedStatus: http.StatusBadGateway,
			expectedCode:   "PROVIDER_ERROR",
			expectedMsg:    "something went wrong on BCV",
		},
		{
			name: "Timeout Provider Error via net.Error maps to 504",
			inputErr: apierrors.NewProviderError(&mockNetError{
				err:     errors.New("connection timeout"),
				timeout: true,
			}),
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   "PROVIDER_ERROR",
			expectedMsg:    "connection timeout",
		},
		{
			name:           "Timeout Provider Error via context.DeadlineExceeded maps to 504",
			inputErr:       apierrors.NewProviderError(context.DeadlineExceeded),
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   "PROVIDER_ERROR",
			expectedMsg:    context.DeadlineExceeded.Error(),
		},
		{
			name:           "Timeout Provider Error via message substring maps to 504",
			inputErr:       apierrors.NewProviderError(errors.New("handshake failed")),
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   "PROVIDER_ERROR",
			expectedMsg:    "handshake failed",
		},
		{
			name:           "Unauthorized error maps to 401",
			inputErr:       apierrors.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
			expectedMsg:    "unauthorized",
		},
		{
			name:           "Unknown internal error maps to 500 and masks message",
			inputErr:       errors.New("some DB issue"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			ctx := context.WithValue(r.Context(), chimw.RequestIDKey, "test-req-id")
			r = r.WithContext(ctx)

			Error(w, r, tt.inputErr)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if got := w.Header().Get("X-Request-ID"); got != "test-req-id" {
				t.Errorf("expected X-Request-ID header %q, got %q", "test-req-id", got)
			}

			var resp ResponseEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.Success {
				t.Error("expected success to be false")
			}

			if resp.Code == nil || string(*resp.Code) != tt.expectedCode {
				gotCode := "nil"
				if resp.Code != nil {
					gotCode = string(*resp.Code)
				}
				t.Errorf("expected error code %q, got %q", tt.expectedCode, gotCode)
			}

			if resp.Error == nil || *resp.Error != tt.expectedMsg {
				gotMsg := "nil"
				if resp.Error != nil {
					gotMsg = *resp.Error
				}
				t.Errorf("expected error message %q, got %q", tt.expectedMsg, gotMsg)
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
		})
	}
}

type mockNetError struct {
	err     error
	timeout bool
}

func (e *mockNetError) Error() string   { return e.err.Error() }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return false }

func TestOKAndCreatedPresenter(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(r.Context(), chimw.RequestIDKey, "test-req-id-ok")
	r = r.WithContext(ctx)

	OK(w, r, "test-data")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Request-ID"); got != "test-req-id-ok" {
		t.Errorf("expected X-Request-ID header %q, got %q", "test-req-id-ok", got)
	}

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if _, exists := raw["request_id"]; exists {
		t.Error("expected 'request_id' key to be omitted from JSON body")
	}
	if raw["data"] != "test-data" {
		t.Errorf("expected data %q, got %v", "test-data", raw["data"])
	}

	// Test Created
	w2 := httptest.NewRecorder()
	Created(w2, r, "created-data")

	if w2.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w2.Code)
	}
	if got := w2.Header().Get("X-Request-ID"); got != "test-req-id-ok" {
		t.Errorf("expected X-Request-ID header %q, got %q", "test-req-id-ok", got)
	}

	var raw2 map[string]any
	if err := json.Unmarshal(w2.Body.Bytes(), &raw2); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if _, exists := raw2["request_id"]; exists {
		t.Error("expected 'request_id' key to be omitted from JSON body")
	}
}
