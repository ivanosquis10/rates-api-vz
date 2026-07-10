package config

import (
	"errors"
	"os"
	"strconv"
)

var (
	// ErrMissingAPIKey is returned when the API_KEY environment variable is not set.
	ErrMissingAPIKey = errors.New("API_KEY environment variable is required")
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	APIKey     string // Required: API key for authentication
	Port       int    // Server port (default: 8080)
	DBPath     string // SQLite database path (default: ./rates.db)
	ScrapeHour int    // Hour to run daily scraping, 0-23 (default: 8)
	RateLimit  int    // Max requests per minute per IP (default: 60)
}

const (
	defaultPort       = 8080
	defaultDBPath     = "./rates.db"
	defaultScrapeHour = 8
	defaultRateLimit  = 60
)

// Load reads configuration from environment variables.
// Returns an error if API_KEY is missing or empty.
func Load() (*Config, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	port := getEnvInt("PORT", defaultPort)
	dbPath := getEnvString("DB_PATH", defaultDBPath)
	scrapeHour := getEnvInt("SCRAPE_HOUR", defaultScrapeHour)
	rateLimit := getEnvInt("RATE_LIMIT", defaultRateLimit)

	return &Config{
		APIKey:     apiKey,
		Port:       port,
		DBPath:     dbPath,
		ScrapeHour: scrapeHour,
		RateLimit:  rateLimit,
	}, nil
}

func getEnvString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
