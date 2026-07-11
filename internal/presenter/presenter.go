package presenter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/ivanosquis10/api-rates-venezuela/internal/apierrors"
	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/httpx"
)

type ResponseEnvelope struct {
	Success   bool             `json:"success"`
	Data      any              `json:"data"`
	ErrorCode *apierrors.Code  `json:"error_code"`
	Error     *string          `json:"error"`
	RequestID string           `json:"request_id"`
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
	httpx.WriteJSON(w, http.StatusOK, ResponseEnvelope{
		Success:   true,
		Data:      data,
		ErrorCode: nil,
		Error:     nil,
		RequestID: reqID,
	})
}

// Created writes an http.StatusCreated enveloped response.
func Created(w http.ResponseWriter, r *http.Request, data any) {
	reqID := httpx.GetRequestID(r.Context())
	httpx.WriteJSON(w, http.StatusCreated, ResponseEnvelope{
		Success:   true,
		Data:      data,
		ErrorCode: nil,
		Error:     nil,
		RequestID: reqID,
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
		default:
			status = http.StatusInternalServerError
		}
	}

	if status == http.StatusInternalServerError {
		slog.Error("internal server error", "error", err, "request_id", reqID)
		message = "internal server error"
	}

	httpx.WriteJSON(w, status, ResponseEnvelope{
		Success:   false,
		ErrorCode: codePtr(code),
		Error:     stringPtr(message),
		RequestID: reqID,
	})
}
