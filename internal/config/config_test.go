package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Set API_KEY (required), leave other vars unset to test defaults
	os.Setenv("API_KEY", "test-key")
	os.Unsetenv("PORT")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("SCRAPE_HOUR")
	os.Unsetenv("RATE_LIMIT")
	defer os.Unsetenv("API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected default Port=8080, got %d", cfg.Port)
	}
	if cfg.DBPath != "./rates.db" {
		t.Errorf("expected default DBPath=./rates.db, got %s", cfg.DBPath)
	}
	if cfg.ScrapeHour != 8 {
		t.Errorf("expected default ScrapeHour=8, got %d", cfg.ScrapeHour)
	}
	if cfg.RateLimit != 60 {
		t.Errorf("expected default RateLimit=60, got %d", cfg.RateLimit)
	}
}

func TestConfigAPIKeyRequired(t *testing.T) {
	// Ensure API_KEY is not set
	os.Unsetenv("API_KEY")
	os.Unsetenv("PORT")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("SCRAPE_HOUR")
	os.Unsetenv("RATE_LIMIT")

	_, err := Load()
	if err == nil {
		t.Error("expected error when API_KEY is missing, got nil")
	}
}

func TestConfigEnvVarOverride(t *testing.T) {
	os.Setenv("API_KEY", "test-key-123")
	os.Setenv("PORT", "3000")
	os.Setenv("DB_PATH", "/tmp/test.db")
	os.Setenv("SCRAPE_HOUR", "14")
	os.Setenv("RATE_LIMIT", "120")
	defer func() {
		os.Unsetenv("API_KEY")
		os.Unsetenv("PORT")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("SCRAPE_HOUR")
		os.Unsetenv("RATE_LIMIT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.APIKey != "test-key-123" {
		t.Errorf("expected APIKey=test-key-123, got %s", cfg.APIKey)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected Port=3000, got %d", cfg.Port)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("expected DBPath=/tmp/test.db, got %s", cfg.DBPath)
	}
	if cfg.ScrapeHour != 14 {
		t.Errorf("expected ScrapeHour=14, got %d", cfg.ScrapeHour)
	}
	if cfg.RateLimit != 120 {
		t.Errorf("expected RateLimit=120, got %d", cfg.RateLimit)
	}
}

func TestConfigEmptyAPIKeyFails(t *testing.T) {
	os.Setenv("API_KEY", "")
	os.Setenv("PORT", "3000")
	defer func() {
		os.Unsetenv("API_KEY")
		os.Unsetenv("PORT")
	}()

	_, err := Load()
	if err == nil {
		t.Error("expected error when API_KEY is empty, got nil")
	}
}

func TestConfigPartialOverride(t *testing.T) {
	os.Setenv("API_KEY", "partial-key")
	os.Setenv("PORT", "9090")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("SCRAPE_HOUR")
	os.Unsetenv("RATE_LIMIT")
	defer func() {
		os.Unsetenv("API_KEY")
		os.Unsetenv("PORT")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("SCRAPE_HOUR")
		os.Unsetenv("RATE_LIMIT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Overridden values
	if cfg.APIKey != "partial-key" {
		t.Errorf("expected APIKey=partial-key, got %s", cfg.APIKey)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", cfg.Port)
	}
	// Default values
	if cfg.DBPath != "./rates.db" {
		t.Errorf("expected default DBPath=./rates.db, got %s", cfg.DBPath)
	}
	if cfg.ScrapeHour != 8 {
		t.Errorf("expected default ScrapeHour=8, got %d", cfg.ScrapeHour)
	}
	if cfg.RateLimit != 60 {
		t.Errorf("expected default RateLimit=60, got %d", cfg.RateLimit)
	}
}

func TestConfigLoadFromDotEnv(t *testing.T) {
	// Ensure no env vars interfere
	os.Unsetenv("API_KEY")
	os.Unsetenv("PORT")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("SCRAPE_HOUR")
	os.Unsetenv("RATE_LIMIT")

	// Create temp dir with .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := []byte("API_KEY=env-file-key\nPORT=9090\n")
	if err := os.WriteFile(envPath, content, 0644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Change working directory to tmpDir so godotenv finds .env
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.APIKey != "env-file-key" {
		t.Errorf("expected APIKey=env-file-key from .env file, got %s", cfg.APIKey)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected Port=9090 from .env file, got %d", cfg.Port)
	}
	// Defaults should apply for unset vars
	if cfg.DBPath != "./rates.db" {
		t.Errorf("expected default DBPath=./rates.db, got %s", cfg.DBPath)
	}
}

func TestConfigEnvVarOverridesDotEnv(t *testing.T) {
	// Set env var that should override .env
	os.Setenv("PORT", "3000")
	os.Unsetenv("API_KEY")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("SCRAPE_HOUR")
	os.Unsetenv("RATE_LIMIT")
	defer os.Unsetenv("PORT")

	// Create temp dir with .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := []byte("API_KEY=env-file-key\nPORT=9090\n")
	if err := os.WriteFile(envPath, content, 0644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Change working directory to tmpDir so godotenv finds .env
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// .env provides API_KEY
	if cfg.APIKey != "env-file-key" {
		t.Errorf("expected APIKey=env-file-key from .env, got %s", cfg.APIKey)
	}
	// Env var overrides .env PORT
	if cfg.Port != 3000 {
		t.Errorf("expected Port=3000 (env var override), got %d", cfg.Port)
	}
}

func TestConfigMissingDotEnvUsesDefaults(t *testing.T) {
	// Ensure no .env file and env vars set for API_KEY only
	os.Setenv("API_KEY", "valid-key")
	os.Unsetenv("PORT")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("SCRAPE_HOUR")
	os.Unsetenv("RATE_LIMIT")
	defer os.Unsetenv("API_KEY")

	// Use a temp dir with no .env file
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected default Port=8080, got %d", cfg.Port)
	}
	if cfg.DBPath != "./rates.db" {
		t.Errorf("expected default DBPath=./rates.db, got %s", cfg.DBPath)
	}
	if cfg.ScrapeHour != 8 {
		t.Errorf("expected default ScrapeHour=8, got %d", cfg.ScrapeHour)
	}
	if cfg.RateLimit != 60 {
		t.Errorf("expected default RateLimit=60, got %d", cfg.RateLimit)
	}
}
