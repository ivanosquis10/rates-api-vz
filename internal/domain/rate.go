package domain

import "time"

// Rate represents an exchange rate entry for Venezuelan currency.
type Rate struct {
	ID        int64     `json:"id"         sqlite:"id"`
	Currency  string    `json:"currency"   sqlite:"currency"`
	Value     float64   `json:"value"      sqlite:"value"`
	ScrapedAt time.Time `json:"scraped_at" sqlite:"scraped_at"`
}
