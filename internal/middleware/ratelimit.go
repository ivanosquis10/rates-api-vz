package middleware

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ivanosquis10/api-rates-venezuela/internal/apierrors"
	"github.com/ivanosquis10/api-rates-venezuela/internal/presenter"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen int64 // Unix timestamp in seconds, read/written atomically via sync/atomic
}

// rateLimiter implements the rate limiting logic and holds the clients state.
type rateLimiter struct {
	mu          sync.RWMutex
	clients     map[string]*client
	limitPerMin int
	pruneAge    time.Duration
}

// RateLimit returns a middleware that limits requests per IP.
// It accepts a context to manage the lifecycle of the background janitor.
func RateLimit(ctx context.Context, limitPerMin int) func(http.Handler) http.Handler {
	if limitPerMin <= 0 {
		limitPerMin = 60
	}

	rl := &rateLimiter{
		clients:     make(map[string]*client),
		limitPerMin: limitPerMin,
		pruneAge:    5 * time.Minute,
	}

	// Janitor goroutine to prune inactive clients
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.prune()
			}
		}
	}()

	return rl.Handler
}

func (rl *rateLimiter) prune() {
	now := time.Now().Unix()
	pruneSecs := int64(rl.pruneAge.Seconds())
	var expiredIPs []string

	rl.mu.RLock()
	for ip, c := range rl.clients {
		if now-atomic.LoadInt64(&c.lastSeen) > pruneSecs {
			expiredIPs = append(expiredIPs, ip)
		}
	}
	rl.mu.RUnlock()

	if len(expiredIPs) > 0 {
		rl.mu.Lock()
		for _, ip := range expiredIPs {
			if c, exists := rl.clients[ip]; exists {
				if time.Now().Unix()-atomic.LoadInt64(&c.lastSeen) > pruneSecs {
					delete(rl.clients, ip)
				}
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		rl.mu.Lock()
		c, exists := rl.clients[ip]
		if !exists {
			c = &client{
				limiter: rate.NewLimiter(rate.Limit(float64(rl.limitPerMin)/60.0), rl.limitPerMin),
			}
			rl.clients[ip] = c
		}
		atomic.StoreInt64(&c.lastSeen, time.Now().Unix())
		rl.mu.Unlock()

		reservation := c.limiter.Reserve()
		if !reservation.OK() {
			presenter.Error(w, r, apierrors.ErrRateLimitExceeded)
			return
		}

		delay := reservation.Delay()
		if delay > 0 {
			reservation.Cancel()
			retryAfter := int(math.Ceil(delay.Seconds()))
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			presenter.Error(w, r, apierrors.ErrRateLimitExceeded)
			return
		}

		next.ServeHTTP(w, r)
	})
}
