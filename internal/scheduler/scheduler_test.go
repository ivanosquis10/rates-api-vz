package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ivanosquis10/api-rates-venezuela/internal/domain"
)

// mockUsecase implements RateScraperUsecase for testing purposes.
type mockUsecase struct {
	calls      int
	scrapeFunc func(ctx context.Context) ([]domain.Rate, error)
}

func (m *mockUsecase) ScrapeRates(ctx context.Context) ([]domain.Rate, error) {
	m.calls++
	return m.scrapeFunc(ctx)
}

func TestScheduler_Initialization(t *testing.T) {
	mu := &mockUsecase{}
	sched := NewScheduler(mu, "0 8 * * *", "*/5 8-18 * * 1-5")
	if sched.usecase != mu {
		t.Error("expected usecase to be set in Scheduler")
	}
	if sched.maintenanceCron != "0 8 * * *" {
		t.Errorf("expected maintenanceCron to be '0 8 * * *', got %s", sched.maintenanceCron)
	}
	if sched.windowCron != "*/5 8-18 * * 1-5" {
		t.Errorf("expected windowCron to be '*/5 8-18 * * 1-5', got %s", sched.windowCron)
	}
	if len(sched.backoffDelays) != 3 {
		t.Errorf("expected 3 backoff delays, got %d", len(sched.backoffDelays))
	}
}

func TestScheduler_StartStop(t *testing.T) {
	mu := &mockUsecase{}
	sched := NewScheduler(mu, "0 8 * * *", "*/5 8-18 * * 1-5")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched.Start(ctx)
	if sched.cron == nil {
		t.Fatal("expected cron to be initialized after Start")
	}

	stopCtx := sched.Stop()
	select {
	case <-stopCtx.Done():
		// Success
	case <-time.After(1 * time.Second):
		t.Error("cron Stop did not complete within 1s")
	}
}

func TestScheduler_RetryLogic_Failure(t *testing.T) {
	mu := &mockUsecase{
		scrapeFunc: func(ctx context.Context) ([]domain.Rate, error) {
			return nil, errors.New("scrape failed")
		},
	}
	sched := NewScheduler(mu, "0 8 * * *", "*/5 8-18 * * 1-5")
	sched.backoffDelays = []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		4 * time.Millisecond,
	}

	sched.executeWithRetry(context.Background())

	if mu.calls != 4 {
		t.Errorf("expected 4 total attempts, got %d", mu.calls)
	}
}

func TestScheduler_RetryLogic_SuccessOnRetry(t *testing.T) {
	calls := 0
	mu := &mockUsecase{
		scrapeFunc: func(ctx context.Context) ([]domain.Rate, error) {
			calls++
			if calls < 3 {
				return nil, errors.New("temporary scrape error")
			}
			return []domain.Rate{
				{Currency: "USD", Value: 36.5},
			}, nil
		},
	}
	sched := NewScheduler(mu, "0 8 * * *", "*/5 8-18 * * 1-5")
	sched.backoffDelays = []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		4 * time.Millisecond,
	}

	sched.executeWithRetry(context.Background())

	if calls != 3 {
		t.Errorf("expected 3 attempts before success, got %d", calls)
	}
}

func TestScheduler_RetryLogic_ContextCancelled(t *testing.T) {
	mu := &mockUsecase{
		scrapeFunc: func(ctx context.Context) ([]domain.Rate, error) {
			return nil, errors.New("scrape failed")
		},
	}
	sched := NewScheduler(mu, "0 8 * * *", "*/5 8-18 * * 1-5")
	sched.backoffDelays = []time.Duration{
		10 * time.Second,
		10 * time.Second,
		10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	sched.executeWithRetry(ctx)
	duration := time.Since(start)

	if duration > 1*time.Second {
		t.Errorf("expected executeWithRetry to exit immediately on cancel, took %v", duration)
	}
	if mu.calls != 1 {
		t.Errorf("expected only 1 call before cancel, got %d", mu.calls)
	}
}
