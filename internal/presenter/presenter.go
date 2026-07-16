package presenter

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ivanosquis10/api-rates-venezuela/internal/apierrors"
	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/http/httpx"
)

type RateResponse struct {
	ID        int64     `json:"id"`
	Currency  string    `json:"currency"`
	Average   float64   `json:"average"`
	UpdatedAt time.Time `json:"updated_at"`
}

func MapToRateResponse(r domain.Rate) RateResponse {
	return RateResponse{
		ID:        r.ID,
		Currency:  r.Currency,
		Average:   r.Value,
		UpdatedAt: r.ScrapedAt,
	}
}

func MapToRateResponses(rates []domain.Rate) []RateResponse {
	res := make([]RateResponse, len(rates))
	for i, r := range rates {
		res[i] = MapToRateResponse(r)
	}
	return res
}

type ResponseEnvelope struct {
	Success bool            `json:"success"`
	Data    any             `json:"data,omitempty"`
	Code    *apierrors.Code `json:"code,omitempty"`
	Error   *string         `json:"error,omitempty"`
}

func codePtr(c apierrors.Code) *apierrors.Code {
	return &c
}

func stringPtr(s string) *string {
	return &s
}

// OK writes an http.StatusOK enveloped response.
func OK(w http.ResponseWriter, r *http.Request, data any) {
	reqID := httpx.GetRequestID(r.Context())
	w.Header().Set("X-Request-ID", reqID)
	httpx.WriteJSON(w, http.StatusOK, ResponseEnvelope{
		Success: true,
		Data:    data,
		Code:    nil,
		Error:   nil,
	})
}

// Created writes an http.StatusCreated enveloped response.
func Created(w http.ResponseWriter, r *http.Request, data any) {
	reqID := httpx.GetRequestID(r.Context())
	w.Header().Set("X-Request-ID", reqID)
	httpx.WriteJSON(w, http.StatusCreated, ResponseEnvelope{
		Success: true,
		Data:    data,
		Code:    nil,
		Error:   nil,
	})
}

// NoContent writes an http.StatusNoContent (204) response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error maps errors and writes an enveloped error response.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	reqID := httpx.GetRequestID(r.Context())

	// Defaults
	code := apierrors.INTERNAL_ERROR
	status := http.StatusInternalServerError
	message := err.Error()

	var apiErr *apierrors.Error

	if errors.Is(err, apierrors.ErrUnauthorized) {
		status = http.StatusUnauthorized
		code = apierrors.UNAUTHORIZED
		message = apierrors.ErrUnauthorized.Message
	} else if errors.Is(err, apierrors.ErrRateLimitExceeded) {
		status = http.StatusTooManyRequests
		code = apierrors.RATE_LIMITED
		message = apierrors.ErrRateLimitExceeded.Message
	} else if errors.Is(err, domain.ErrNotFound) {
		status = http.StatusNotFound
		code = apierrors.NOT_FOUND
		message = domain.ErrNotFound.Error()
	} else if errors.Is(err, domain.ErrInvalidInput) {
		status = http.StatusBadRequest
		code = apierrors.BAD_REQUEST
		message = domain.ErrInvalidInput.Error()
	} else if errors.Is(err, domain.ErrDuplicateRate) {
		status = http.StatusBadRequest
		code = apierrors.BAD_REQUEST
		message = domain.ErrDuplicateRate.Error()
	} else if errors.As(err, &apiErr) {
		code = apiErr.Code
		message = apiErr.Message
		switch code {
		case apierrors.UNAUTHORIZED:
			status = http.StatusUnauthorized
		case apierrors.RATE_LIMITED:
			status = http.StatusTooManyRequests
		case apierrors.NOT_FOUND:
			status = http.StatusNotFound
		case apierrors.BAD_REQUEST:
			status = http.StatusBadRequest
		case apierrors.PROVIDER_ERROR:
			isTimeout := false
			if apiErr.Err != nil {
				var netErr net.Error
				if errors.As(apiErr.Err, &netErr) && netErr.Timeout() {
					isTimeout = true
				} else if errors.Is(apiErr.Err, context.DeadlineExceeded) {
					isTimeout = true
				}
			}
			if !isTimeout {
				lowerMsg := strings.ToLower(apiErr.Message)
				if strings.Contains(lowerMsg, "timeout") ||
					strings.Contains(lowerMsg, "deadline") ||
					strings.Contains(lowerMsg, "handshake") {
					isTimeout = true
				}
				if apiErr.Err != nil {
					lowerErr := strings.ToLower(apiErr.Err.Error())
					if strings.Contains(lowerErr, "timeout") ||
						strings.Contains(lowerErr, "deadline") ||
						strings.Contains(lowerErr, "handshake") {
						isTimeout = true
					}
				}
			}

			if isTimeout {
				status = http.StatusGatewayTimeout
			} else {
				status = http.StatusBadGateway
			}
		default:
			status = http.StatusInternalServerError
		}
	}

	if status == http.StatusInternalServerError {
		slog.Error("internal server error", "error", err, "request_id", reqID)
		message = "internal server error"
	}

	w.Header().Set("X-Request-ID", reqID)
	httpx.WriteJSON(w, status, ResponseEnvelope{
		Success: false,
		Code:    codePtr(code),
		Error:   stringPtr(message),
	})
}
