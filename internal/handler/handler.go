package handler

import (
	"context"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/ivanosquis10/api-rates-venezuela/internal/usecase"
)

// Usecaser is the interface that RateUsecase satisfies.
// Used for dependency injection and testability.
type Usecaser interface {
	GetLatestRate(ctx context.Context, currency string) (domain.Rate, error)
	GetHistoryRates(ctx context.Context, currency, from, to string, limit int) ([]domain.Rate, error)
	ScrapeRates(ctx context.Context) ([]domain.Rate, error)
}

// Handler holds the usecase dependency for all HTTP handlers.
type Handler struct {
	uc Usecaser
}

// NewHandler creates a Handler with the given concrete usecase.
func NewHandler(uc *usecase.RateUsecase) *Handler {
	return &Handler{uc: uc}
}

// NewHandlerFromUsecaser creates a Handler from any Usecaser implementation.
// Primarily used for testing with mock usecases.
func NewHandlerFromUsecaser(uc Usecaser) *Handler {
	return &Handler{uc: uc}
}
