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

	// Custom middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(deps.RateLimiter.Handler)
	r.Use(middleware.Auth(deps.Config.APIKey))

	// Custom NotFound handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		presenter.Error(w, r, apierrors.New(apierrors.NOT_FOUND, "requested endpoint does not exist"))
	})

	// Routes
	r.Get("/rates", deps.Handler.GetRates)
	r.Get("/rates/history", deps.Handler.GetHistory)
	r.Route("/admin", func(r chi.Router) {
		r.Post("/scrape", deps.Handler.TriggerScrape)
	})

	return r
}
