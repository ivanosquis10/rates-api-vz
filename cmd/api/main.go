package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/ivanosquis10/api-rates-venezuela/internal/config"
	"github.com/ivanosquis10/api-rates-venezuela/internal/handler"
	"github.com/ivanosquis10/api-rates-venezuela/internal/http/router"
	"github.com/ivanosquis10/api-rates-venezuela/internal/middleware"
	"github.com/ivanosquis10/api-rates-venezuela/internal/scheduler"
	"github.com/ivanosquis10/api-rates-venezuela/internal/scraper"
	"github.com/ivanosquis10/api-rates-venezuela/internal/store"
	"github.com/ivanosquis10/api-rates-venezuela/internal/usecase"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Open database
	db, err := store.New(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Setup custom client with timeout settings
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSHandshakeTimeout = 10 * time.Second
	client := &http.Client{
		Transport: customTransport,
		Timeout:   15 * time.Second,
	}

	// Wire dependencies: store → scraper → usecase → handler
	repo := store.NewStore(db)
	bcvScraper := scraper.NewBCVScraper(
		client,
		"https://www.bcv.org.ve",
	)
	uc := usecase.NewRateUsecase(repo, bcvScraper)
	h := handler.NewHandler(uc)

	sched := scheduler.NewScheduler(uc, cfg.ScrapeHour)
	sched.Start(ctx)

	rl := middleware.NewRateLimiter(ctx, cfg.RateLimit)

	// Build router engine
	engine := router.New(router.Deps{
		Handler:     h,
		Config:      cfg,
		RateLimiter: rl,
	})

	// Start server with graceful shutdown
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// Listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("server shutting down")
	cancel()

	slog.Info("stopping scheduler")
	<-sched.Stop().Done()
	slog.Info("scheduler stopped")
}
