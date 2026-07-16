package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	// ErrMissingAPIKey is returned when the API_KEY environment variable is not set.
	ErrMissingAPIKey = errors.New("API_KEY environment variable is required")
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	APIKey                string // Required: API key for authentication
	Port                  int    // Server port (default: 8080)
	DBPath                string // SQLite database path (default: ./rates.db)
	ScrapeCronMaintenance string // Cron expression for main scrape job (default: "0 8 * * *")
	ScrapeCronWindow      string // Cron expression for window scrape job (default: "*/5 8-18 * * 1-5")
	RateLimit             int    // Max requests per minute per IP (default: 60)
}

const (
	defaultPort                   = 8080
	defaultDBPath                 = "./rates.db"
	defaultScrapeCronMaintenance = "0 8 * * *"
	defaultScrapeCronWindow      = "*/5 8-18 * * 1-5"
	defaultRateLimit              = 60
)

// Load reads configuration from environment variables.
// It first attempts to load a .env file (non-fatal if missing),
// then allows environment variable overrides.
// Returns an error if API_KEY is missing or empty.
func Load() (*Config, error) {
	// Load .env file if present (ignore error if missing)
	_ = godotenv.Load()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	port := getEnvInt("PORT", defaultPort)
	dbPath := getEnvString("DB_PATH", defaultDBPath)
	scrapeCronMaintenance := getEnvString("SCRAPE_CRON_MAINTENANCE", defaultScrapeCronMaintenance)
	scrapeCronWindow := getEnvString("SCRAPE_CRON_WINDOW", defaultScrapeCronWindow)
	rateLimit := getEnvInt("RATE_LIMIT", defaultRateLimit)

	return &Config{
		APIKey:                apiKey,
		Port:                  port,
		DBPath:                dbPath,
		ScrapeCronMaintenance: scrapeCronMaintenance,
		ScrapeCronWindow:      scrapeCronWindow,
		RateLimit:             rateLimit,
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
