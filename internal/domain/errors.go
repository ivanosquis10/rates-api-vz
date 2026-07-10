package domain

import "errors"

var (
	// ErrNotFound indicates a requested rate was not found.
	ErrNotFound = errors.New("rate not found")

	// ErrDuplicateRate indicates an attempt to insert a duplicate rate.
	ErrDuplicateRate = errors.New("duplicate rate")

	// ErrInvalidInput indicates invalid input data.
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabase indicates a database operation failed.
	ErrDatabase = errors.New("database error")

	// ErrScrapeFailed indicates a scraper HTTP request failed.
	ErrScrapeFailed = errors.New("scrape failed")

	// ErrParseFailed indicates HTML parsing of scraped content failed.
	ErrParseFailed = errors.New("parse failed")
)
