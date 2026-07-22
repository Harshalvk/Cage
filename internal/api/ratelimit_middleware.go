package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/harshalvk/cage/internal/auth"
	"github.com/harshalvk/cage/internal/ratelimit"
)

func RateLimitMiddleware(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			rawKey := strings.TrimPrefix(authHeader, "Bearer ")
			keyHash := auth.HashKey(rawKey)

			allowed, remaining, err := limiter.Allow(r.Context(), keyHash)
			if err != nil {
				// Fail open — a Redis outage shouldn't take down the whole API.
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", remaining))

			if !allowed {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
