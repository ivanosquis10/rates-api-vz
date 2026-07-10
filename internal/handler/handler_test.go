package handler

import (
	"testing"

	"github.com/ivanosquis10/api-rates-venezuela/internal/usecase"
)

func TestNewHandler_NonNil(t *testing.T) {
	// NewHandler with nil usecase should still return a Handler (defensive).
	h := NewHandler(nil)
	if h == nil {
		t.Fatal("expected non-nil Handler from NewHandler(nil)")
	}
}

func TestNewHandler_WithUsecase(t *testing.T) {
	// NewHandler with a real usecase should return a Handler wired to it.
	uc := &usecase.RateUsecase{}
	h := NewHandler(uc)
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
	if h.uc != uc {
		t.Error("expected Handler.uc to match the injected usecase")
	}
}
