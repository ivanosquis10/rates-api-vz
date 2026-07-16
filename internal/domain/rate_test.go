package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRateJSONSerialization(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

	r := Rate{
		ID:        1,
		Currency:  "USD",
		Value:     36.50,
		ScrapedAt: now,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded["currency"] != "USD" {
		t.Errorf("expected currency=USD, got %v", decoded["currency"])
	}
	if decoded["value"] != 36.5 {
		t.Errorf("expected value=36.5, got %v", decoded["value"])
	}
}

func TestRateStructFields(t *testing.T) {
	now := time.Now()
	r := Rate{
		ID:        1,
		Currency:  "EUR",
		Value:     42.0,
		ScrapedAt: now,
	}

	if r.ID != 1 {
		t.Errorf("expected ID=1, got %d", r.ID)
	}
	if r.Currency != "EUR" {
		t.Errorf("expected Currency=EUR, got %s", r.Currency)
	}
	if r.Value != 42.0 {
		t.Errorf("expected Value=42.0, got %f", r.Value)
	}
	if !r.ScrapedAt.Equal(now) {
		t.Errorf("expected ScrapedAt=%v, got %v", now, r.ScrapedAt)
	}
}
