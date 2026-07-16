package domain

import "context"

// Repository defines the persistence contract for rate data.
type Repository interface {
	// SaveRates persists a batch of rates atomically.
	SaveRates(ctx context.Context, rates []Rate) error

	// GetLatestRate returns the most recent rate for the given currency.
	GetLatestRate(ctx context.Context, currency string) (Rate, error)

	// GetHistoryRates returns rates matching the given filters, ordered by scraped_at DESC,
	// limited to the specified count. Returns empty slice (not nil) when no rates match.
	GetHistoryRates(ctx context.Context, currency, from, to string, limit int) ([]Rate, error)
}
