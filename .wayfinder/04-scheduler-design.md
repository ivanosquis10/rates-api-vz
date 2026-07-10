# Ticket 04: Scheduler Design

**Type**: Grilling (HITL)
**Blocked by**: Ticket 01 (scraping library)
**Status**: RESOLVED

## Question

How should the daily scraping job be scheduled and executed?

## Resolution

### Chosen approach: `robfig/cron` + goroutine worker

```go
// scheduler/scheduler.go
package scheduler

import (
    "github.com/robfig/cron/v3"
    // ...
)

type Scheduler struct {
    cron    *cron.Cron
    usecase RateUseCase
}

func New(hour int, uc RateUseCase) *Scheduler {
    c := cron.New(cron.WithLocation(caracasTZ))
    s := &Scheduler{cron: c, usecase: uc}
    
    // Schedule daily scraping at specified hour
    spec := fmt.Sprintf("0 %d * * *", hour) // e.g., "0 8 * * *"
    c.AddFunc(spec, s.scrapeWithRetry)
    
    return s
}

func (s *Scheduler) Start()      { s.cron.Start() }
func (s *Scheduler) Stop()       { s.cron.Stop() }
func (s *Scheduler) TriggerNow() { go s.scrapeWithRetry() }
```

### Retry strategy: 3 attempts + exponential backoff

```go
func (s *Scheduler) scrapeWithRetry() {
    maxRetries := 3
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := s.usecase.ScrapeRates(context.Background())
        if err == nil {
            log.Printf("Scrape successful on attempt %d", attempt+1)
            return
        }
        
        wait := time.Duration(1<<uint(attempt)) * time.Minute // 1m, 2m, 4m
        log.Printf("Scrape failed (attempt %d/%d): %v. Retrying in %v", 
            attempt+1, maxRetries, err, wait)
        time.Sleep(wait)
    }
    log.Printf("Scrape failed after %d attempts", maxRetries)
}
```

### Manual trigger endpoint

```
POST /admin/scrape
Header: X-API-Key: <key>
Response: { "status": "scrape triggered" }
```

This endpoint calls `scheduler.TriggerNow()` — executes scraping in a goroutine, returns immediately.

### Timezone handling

BCV publishes in Caracas time (UTC-4). Schedule in `America/Caracas` timezone:

```go
loc, _ := time.LoadLocation("America/Caracas")
c := cron.New(cron.WithLocation(loc))
```

### Key decisions

| Decision | Chosen | Why |
|----------|--------|-----|
| Library | robfig/cron | Cron expressions, timezone support, mature |
| Retries | 3 + exponential | Balance between reliability and not hammering the site |
| Manual trigger | POST /admin/scrape | Useful for debugging and initial testing |
| Timezone | America/Caracas | BCV publishes in local time |
| Concurrency | Single goroutine | Prevent overlapping scrapes |
