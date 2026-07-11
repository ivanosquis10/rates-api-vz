package middleware

import (
	"fmt"
	"net/http"

	"github.com/ivanosquis10/api-rates-venezuela/internal/presenter"
)

// Recovery returns middleware that catches panics and returns HTTP 500.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				presenter.Error(w, r, fmt.Errorf("panic recovered: %v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
