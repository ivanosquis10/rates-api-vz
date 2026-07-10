package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ivanosquis10/api-rates-venezuela/internal/config"
	"github.com/ivanosquis10/api-rates-venezuela/internal/handler"
	"github.com/ivanosquis10/api-rates-venezuela/internal/middleware"
	"github.com/ivanosquis10/api-rates-venezuela/internal/scraper"
	"github.com/ivanosquis10/api-rates-venezuela/internal/store"
	"github.com/ivanosquis10/api-rates-venezuela/internal/usecase"
)

func main() {
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

	// Wire dependencies: store → scraper → usecase → handler
	repo := store.NewStore(db)
	bcvScraper := scraper.NewBCVScraper(
		http.DefaultClient,
		"https://www.bcv.org.ve",
		"https://www.bcv.org.ve/estadisticas",
	)
	uc := usecase.NewRateUsecase(repo, bcvScraper)
	h := handler.NewHandler(uc)

	// Build Chi router
	r := chi.NewRouter()

	// Chi's built-in middleware
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)

	// Custom middleware: recovery first, then logging
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)

	// Routes
	r.Get("/rates", h.GetRates)
	r.Get("/rates/history", h.GetHistory)
	r.Route("/admin", func(r chi.Router) {
		r.Post("/scrape", h.TriggerScrape)
	})

	// Start server with graceful shutdown
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
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
}
