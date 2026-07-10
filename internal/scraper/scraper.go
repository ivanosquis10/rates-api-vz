package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

// Scraper defines the contract for exchange rate data sources.
type Scraper interface {
	Scrape(ctx context.Context) ([]domain.Rate, error)
}

// BCVScraper fetches and parses rates from the Banco Central de Venezuela.
type BCVScraper struct {
	client      *http.Client
	homepageURL string
	statsURL    string
}

// Selector constants — isolated for easy update when BCV changes HTML.
const (
	selUSDRate  = "#dolar .strong-tb"
	selEURRate  = "#euro .strong-tb"
	selBankTable = ".views-table tbody tr"
	selDate     = ".date-display-single"
	selDateAttr = "content"
)

// NewBCVScraper creates a new BCVScraper with the given HTTP client and URLs.
func NewBCVScraper(client *http.Client, homepageURL, statsURL string) *BCVScraper {
	return &BCVScraper{
		client:      client,
		homepageURL: homepageURL,
		statsURL:    statsURL,
	}
}

// Scrape fetches BCV homepage and statistics page, parses exchange rates,
// and returns all discovered rates with a shared ScrapedAt timestamp.
func (s *BCVScraper) Scrape(ctx context.Context) ([]domain.Rate, error) {
	// Fetch homepage for reference rates
	homeDoc, err := s.fetchPage(ctx, s.homepageURL)
	if err != nil {
		return nil, fmt.Errorf("fetch reference page: %w", err)
	}

	// Parse reference rates (USD, EUR)
	refRates, err := parseReferenceRates(homeDoc)
	if err != nil {
		return nil, fmt.Errorf("parse reference rates: %w", err)
	}

	// Fetch statistics page for bank rates and date
	statsDoc, err := s.fetchPage(ctx, s.statsURL)
	if err != nil {
		return nil, fmt.Errorf("fetch statistics page: %w", err)
	}

	// Parse date
	scrapedAt, err := parseDate(statsDoc)
	if err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}

	// Parse bank rates
	bankRates, err := parseBankRates(statsDoc)
	if err != nil {
		return nil, fmt.Errorf("parse bank rates: %w", err)
	}

	// Combine all rates and set ScrapedAt
	allRates := make([]domain.Rate, 0, len(refRates)+len(bankRates))
	allRates = append(allRates, refRates...)
	allRates = append(allRates, bankRates...)

	for i := range allRates {
		allRates[i].ScrapedAt = scrapedAt
	}

	return allRates, nil
}

// fetchPage performs an HTTP GET, checks for 2xx status, and parses the response as HTML.
func (s *BCVScraper) fetchPage(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return doc, nil
}

// parseReferenceRates extracts USD and EUR reference rates from the BCV homepage.
func parseReferenceRates(doc *goquery.Document) ([]domain.Rate, error) {
	var rates []domain.Rate

	usdVal, err := extractNumeric(doc, selUSDRate)
	if err != nil {
		return nil, fmt.Errorf("parse USD rate: %w", err)
	}
	rates = append(rates, domain.Rate{
		Currency: "USD",
		RateType: "reference",
		Value:    usdVal,
	})

	eurVal, err := extractNumeric(doc, selEURRate)
	if err != nil {
		return nil, fmt.Errorf("parse EUR rate: %w", err)
	}
	rates = append(rates, domain.Rate{
		Currency: "EUR",
		RateType: "reference",
		Value:    eurVal,
	})

	return rates, nil
}

// parseBankRates extracts buy/sell rates from the bank rates table.
func parseBankRates(doc *goquery.Document) ([]domain.Rate, error) {
	var rates []domain.Rate

	doc.Find(selBankTable).Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 3 {
			return // skip malformed rows
		}

		bankName := strings.TrimSpace(cells.Eq(0).Text())
		if bankName == "" {
			return // skip rows without bank name
		}

		buyVal, err := parseNumericText(cells.Eq(1).Text())
		if err != nil {
			return // skip rows with invalid buy rate
		}

		sellVal, err := parseNumericText(cells.Eq(2).Text())
		if err != nil {
			return // skip rows with invalid sell rate
		}

		rates = append(rates, domain.Rate{
			Currency: "USD",
			RateType: "buy",
			Bank:     bankName,
			Value:    buyVal,
		})
		rates = append(rates, domain.Rate{
			Currency: "USD",
			RateType: "sell",
			Bank:     bankName,
			Value:    sellVal,
		})
	})

	return rates, nil
}

// parseDate extracts the ISO 8601 publication date from the statistics page.
func parseDate(doc *goquery.Document) (time.Time, error) {
	dateStr, exists := doc.Find(selDate).Attr(selDateAttr)
	if !exists || strings.TrimSpace(dateStr) == "" {
		return time.Time{}, fmt.Errorf("date element not found or empty")
	}

	t, err := time.Parse(time.RFC3339, strings.TrimSpace(dateStr))
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", dateStr, err)
	}

	return t, nil
}

// extractNumeric finds an element by selector and parses its text as float64.
func extractNumeric(doc *goquery.Document, selector string) (float64, error) {
	text := strings.TrimSpace(doc.Find(selector).Text())
	if text == "" {
		return 0, fmt.Errorf("element %q not found or empty", selector)
	}
	return parseNumericText(text)
}

// parseNumericText parses a string as float64, handling comma decimal separators.
func parseNumericText(text string) (float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("empty numeric value")
	}
	// Handle comma as decimal separator (common in Venezuelan formatting)
	text = strings.ReplaceAll(text, ",", ".")
	val, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %q as number: %w", text, err)
	}
	return val, nil
}
