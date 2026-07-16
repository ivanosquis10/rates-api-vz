package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/ivanosquis10/api-rates-venezuela/internal/apierrors"
	"github.com/ivanosquis10/api-rates-venezuela/internal/config"
	"github.com/ivanosquis10/api-rates-venezuela/internal/handler"
	"github.com/ivanosquis10/api-rates-venezuela/internal/middleware"
	"github.com/ivanosquis10/api-rates-venezuela/internal/presenter"
)

// Deps holds router dependencies
type Deps struct {
	Handler     *handler.Handler
	Config      *config.Config
	RateLimiter *middleware.RateLimiter
}

// New constructs the decoupled chi router
func New(deps Deps) http.Handler {
	r := chi.NewRouter()

	// Chi's built-in middleware
	r.Use(chimw.ClientIPFromRemoteAddr)
	r.Use(chimw.RequestID)
	r.Use(middleware.CORS)

	// Custom middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(deps.RateLimiter.Handler)
	r.Use(middleware.Auth(deps.Config.APIKey))

	// Custom NotFound handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		presenter.Error(w, r, apierrors.New(apierrors.NOT_FOUND, "requested endpoint does not exist"))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", deps.Handler.Health)
		r.Get("/dollars", deps.Handler.GetUSD)
		r.Get("/dollars/official", deps.Handler.GetOfficialUSD)
		r.Get("/euros", deps.Handler.GetEUR)
		r.Get("/euros/official", deps.Handler.GetOfficialEUR)
		r.Get("/history/dollars", deps.Handler.GetUSDHistory)
		r.Get("/history/euros", deps.Handler.GetEURHistory)
		r.Route("/admin", func(r chi.Router) {
			r.Post("/scrape", deps.Handler.TriggerScrape)
		})
	})

	return r
}
