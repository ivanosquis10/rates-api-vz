package presenter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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

			Error(w, r, tt.inputErr)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
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
