package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	_ "time/tzdata"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
	"github.com/robfig/cron/v3"
)

// RateScraperUsecase defines the usecase dependency for the scheduler.
type RateScraperUsecase interface {
	ScrapeRates(ctx context.Context) ([]domain.Rate, error)
}

// Scheduler manages the daily cron job for scraping exchange rates.
type Scheduler struct {
	cron          *cron.Cron
	usecase       RateScraperUsecase
	scrapeHour    int
	backoffDelays []time.Duration
}

// NewScheduler creates a Scheduler with the given usecase and scrape hour.
func NewScheduler(uc RateScraperUsecase, scrapeHour int) *Scheduler {
	return &Scheduler{
		usecase:    uc,
		scrapeHour: scrapeHour,
		backoffDelays: []time.Duration{
			1 * time.Minute,
			2 * time.Minute,
			4 * time.Minute,
		},
	}
}

// Start starts the daily cron job using the America/Caracas timezone.
func (s *Scheduler) Start(ctx context.Context) {
	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		slog.Warn("failed to load America/Caracas timezone, defaulting to UTC", "error", err)
		loc = time.UTC
	}

	s.cron = cron.New(cron.WithLocation(loc))
	cronExpr := fmt.Sprintf("0 %d * * *", s.scrapeHour)
	_, err = s.cron.AddFunc(cronExpr, func() {
		s.executeWithRetry(ctx)
	})
	if err != nil {
		slog.Error("failed to register daily scraping cron job", "expr", cronExpr, "error", err)
	}

	s.cron.Start()
	slog.Info("scheduler started", "scrape_hour", s.scrapeHour, "timezone", loc.String())
}

// RunNow executes an immediate scrape with retry, reusing the existing
// backoff logic. It does not block the scheduler lifecycle.
func (s *Scheduler) RunNow(ctx context.Context) {
	slog.Info("running initial scrape at startup")
	s.executeWithRetry(ctx)
	slog.Info("initial scrape completed")
}

// Stop stops the scheduler and returns a context that is closed when running jobs finish.
func (s *Scheduler) Stop() context.Context {
	if s.cron != nil {
		return s.cron.Stop()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// executeWithRetry runs ScrapeRates with up to 4 total attempts.
func (s *Scheduler) executeWithRetry(ctx context.Context) {
	maxAttempts := 4
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		startTime := time.Now()
		rates, err := s.usecase.ScrapeRates(ctx)
		elapsed := time.Since(startTime).Milliseconds()

		if err == nil {
			for _, rate := range rates {
				slog.Info("rate scraped successfully",
					"duration_ms", elapsed,
					"currency", rate.Currency,
					"value", rate.Value,
				)
			}
			return
		}

		slog.Error("scrape attempt failed",
			"error", err,
			"attempt", attempt,
		)

		if attempt == maxAttempts {
			return
		}

		delay := s.backoffDelays[attempt-1]
		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			timer.Stop()
		}
	}
}
