package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", name, err)
	}
	return string(data)
}

func TestScrapeSuccess(t *testing.T) {
	homepageHTML := loadFixture(t, "bcv-homepage.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(homepageHTML))
	}))
	defer server.Close()

	scraper := NewBCVScraper(server.Client(), server.URL+"/")

	rates, err := scraper.Scrape(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rates) < 2 {
		t.Fatalf("expected at least 2 rates (USD + EUR), got %d", len(rates))
	}

	// Find USD reference rate
	var usdRate *domain.Rate
	for i := range rates {
		if rates[i].Currency == "USD" && rates[i].RateType == "reference" {
			usdRate = &rates[i]
			break
		}
	}
	if usdRate == nil {
		t.Fatal("USD reference rate not found")
	}
	if usdRate.Value != 36.50 {
		t.Errorf("USD rate = %v, want 36.50", usdRate.Value)
	}
	if usdRate.Bank != "" {
		t.Errorf("USD reference rate bank = %q, want empty", usdRate.Bank)
	}

	// Find EUR reference rate
	var eurRate *domain.Rate
	for i := range rates {
		if rates[i].Currency == "EUR" && rates[i].RateType == "reference" {
			eurRate = &rates[i]
			break
		}
	}
	if eurRate == nil {
		t.Fatal("EUR reference rate not found")
	}
	if eurRate.Value != 38.20 {
		t.Errorf("EUR rate = %v, want 38.20", eurRate.Value)
	}

	// Find bank rates
	var bankRates []domain.Rate
	for _, r := range rates {
		if r.RateType == "buy" || r.RateType == "sell" {
			bankRates = append(bankRates, r)
		}
	}
	if len(bankRates) != 4 {
		t.Errorf("expected 4 bank rates (2 banks × buy/sell), got %d", len(bankRates))
	}

	// Check ScrapedAt is set
	expectedDate := time.Date(2026, 7, 10, 0, 0, 0, 0, time.FixedZone("VET", -4*60*60))
	for _, r := range rates {
		if !r.ScrapedAt.Equal(expectedDate) {
			t.Errorf("rate %s %s: ScrapedAt = %v, want %v", r.Currency, r.RateType, r.ScrapedAt, expectedDate)
		}
	}
}

func TestScrapeMissingUSD(t *testing.T) {
	html := `<html><body>
		<div id="euro"><span class="strong-tb">38.20</span></div>
	</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	scraper := NewBCVScraper(server.Client(), server.URL+"/")

	_, err := scraper.Scrape(context.Background())
	if err == nil {
		t.Fatal("expected error for missing USD selector, got nil")
	}
	if !strings.Contains(err.Error(), "USD") {
		t.Errorf("error should mention USD, got: %v", err)
	}
}

func TestScrapeNonNumericValue(t *testing.T) {
	html := `<html><body>
		<div id="dolar"><span class="strong-tb">N/A</span></div>
		<div id="euro"><span class="strong-tb">38.20</span></div>
	</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	scraper := NewBCVScraper(server.Client(), server.URL+"/")

	_, err := scraper.Scrape(context.Background())
	if err == nil {
		t.Fatal("expected error for non-numeric USD value, got nil")
	}
	if !strings.Contains(err.Error(), "USD") {
		t.Errorf("error should mention USD, got: %v", err)
	}
}

func TestScrapeEmptyBankTable(t *testing.T) {
	homepageHTML := `<html><body>
		<div id="dolar"><span class="strong-tb">36.50</span></div>
		<div id="euro"><span class="strong-tb">38.20</span></div>
		<span class="date-display-single" content="2026-07-10T00:00:00-04:00">10 julio 2026</span>
		<table class="views-table"><tbody></tbody></table>
	</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(homepageHTML))
	}))
	defer server.Close()

	scraper := NewBCVScraper(server.Client(), server.URL+"/")

	rates, err := scraper.Scrape(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have exactly 2 reference rates (USD + EUR), no bank rates
	if len(rates) != 2 {
		t.Errorf("expected 2 rates (references only), got %d", len(rates))
	}
	for _, r := range rates {
		if r.RateType != "reference" {
			t.Errorf("expected only reference rates, got rate_type=%q", r.RateType)
		}
	}
}

func TestScrapeMalformedHTML(t *testing.T) {
	html := `<html><body><div id="dolar"><span class="strong-tb">36.50</span></div><div id="euro"><span class="strong-tb">38.20</span></div></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	scraper := NewBCVScraper(server.Client(), server.URL+"/")

	// goquery is lenient with malformed HTML, so this tests that the scraper
	// doesn't panic and handles the parse gracefully
	rates, err := scraper.Scrape(context.Background())
	if err != nil {
		// Error is acceptable for malformed content
		return
	}
	// If no error, rates should still be valid
	if len(rates) < 2 {
		t.Errorf("expected at least 2 reference rates from malformed HTML, got %d", len(rates))
	}
}

func TestScrapeHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	scraper := NewBCVScraper(server.Client(), server.URL+"/")

	_, err := scraper.Scrape(context.Background())
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}

func TestScrapeContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handler should not be reached if context is cancelled
		t.Error("handler called despite cancelled context")
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	scraper := NewBCVScraper(server.Client(), server.URL+"/")

	_, err := scraper.Scrape(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
