package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// Auth returns a middleware that validates requests using an API key.
// It compares the X-API-Key header against the configured apiKey using
// a timing-attack resistant comparison of their SHA-256 hashes.
func Auth(apiKey string) func(http.Handler) http.Handler {
	keyHash := sha256.Sum256([]byte(apiKey))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inputKey := r.Header.Get("X-API-Key")
			if inputKey == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"unauthorized"}}`))
				return
			}

			inputHash := sha256.Sum256([]byte(inputKey))

			if subtle.ConstantTimeCompare(keyHash[:], inputHash[:]) != 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"unauthorized"}}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
