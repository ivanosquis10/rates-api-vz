package main

import (
	"os"
	"testing"

	"github.com/ivanosquis10/api-rates-venezuela/internal/config"
	"github.com/ivanosquis10/api-rates-venezuela/internal/store"
)

func TestAppWiring(t *testing.T) {
	// Set up required env
	os.Setenv("API_KEY", "test-key")
	defer os.Unsetenv("API_KEY")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Initialize DB
	db, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("store.New(:memory:) failed: %v", err)
	}
	defer db.Close()

	// Verify DB is usable
	if err := db.Ping(); err != nil {
		t.Errorf("expected pingable DB, got error: %v", err)
	}

	// Store satisfies repository interface
	s := store.NewStore(db)
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}
