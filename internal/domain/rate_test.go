package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRateJSONSerialization(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

	// Bank-specific rate
	r := Rate{
		ID:        1,
		Currency:  "USD",
		RateType:  "reference",
		Bank:      "Banesco",
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
	if decoded["rate_type"] != "reference" {
		t.Errorf("expected rate_type=reference, got %v", decoded["rate_type"])
	}
	if decoded["bank"] != "Banesco" {
		t.Errorf("expected bank=Banesco, got %v", decoded["bank"])
	}
	if decoded["value"] != 36.5 {
		t.Errorf("expected value=36.5, got %v", decoded["value"])
	}

	// Reference rate with empty Bank
	ref := Rate{
		ID:        2,
		Currency:  "USD",
		RateType:  "reference",
		Bank:      "",
		Value:     36.50,
		ScrapedAt: now,
	}

	refData, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("json.Marshal failed for ref rate: %v", err)
	}

	var refDecoded map[string]interface{}
	if err := json.Unmarshal(refData, &refDecoded); err != nil {
		t.Fatalf("json.Unmarshal failed for ref rate: %v", err)
	}

	// Bank should appear as empty string in JSON (not null)
	if refDecoded["bank"] != "" {
		t.Errorf("expected bank=\"\", got %v", refDecoded["bank"])
	}
}

func TestRateStructFields(t *testing.T) {
	now := time.Now()
	r := Rate{
		ID:        1,
		Currency:  "EUR",
		RateType:  "parallel",
		Bank:      "",
		Value:     42.0,
		ScrapedAt: now,
	}

	if r.ID != 1 {
		t.Errorf("expected ID=1, got %d", r.ID)
	}
	if r.Currency != "EUR" {
		t.Errorf("expected Currency=EUR, got %s", r.Currency)
	}
	if r.RateType != "parallel" {
		t.Errorf("expected RateType=parallel, got %s", r.RateType)
	}
	if r.Bank != "" {
		t.Errorf("expected Bank=\"\", got %s", r.Bank)
	}
	if r.Value != 42.0 {
		t.Errorf("expected Value=42.0, got %f", r.Value)
	}
	if !r.ScrapedAt.Equal(now) {
		t.Errorf("expected ScrapedAt=%v, got %v", now, r.ScrapedAt)
	}
}
