package router

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/ivanosquis10/api-rates-venezuela/internal/config"
	"github.com/ivanosquis10/api-rates-venezuela/internal/handler"
	"github.com/ivanosquis10/api-rates-venezuela/internal/middleware"
)

// Deps holds router dependencies
type Deps struct {
	Handler *handler.Handler
	Config  *config.Config
	Context context.Context
}

// New constructs the decoupled chi router
func New(deps Deps) http.Handler {
	r := chi.NewRouter()

	// Chi's built-in middleware
	r.Use(chimw.ClientIPFromRemoteAddr)
	r.Use(chimw.RequestID)

	// Custom middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.RateLimit(deps.Context, deps.Config.RateLimit))
	r.Use(middleware.Auth(deps.Config.APIKey))

	// Routes
	r.Get("/rates", deps.Handler.GetRates)
	r.Get("/rates/history", deps.Handler.GetHistory)
	r.Route("/admin", func(r chi.Router) {
		r.Post("/scrape", deps.Handler.TriggerScrape)
	})

	return r
}
