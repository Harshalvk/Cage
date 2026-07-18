package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/harshalvk/cage/internal/auth"
	"github.com/harshalvk/cage/internal/cache"
	"github.com/harshalvk/cage/internal/store"
)

const apiKeyCacheTTL = 5 * time.Minute

func (a *API) AuthMiddleware(st *store.Store, c *cache.Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			rawKey := strings.TrimPrefix(authHeader, "Bearer ")
			keyHash := auth.HashKey(rawKey)

			valid, err := validateWithCache(r.Context(), st, c, keyHash)
			if err != nil {
				http.Error(w, "failed to validate api key", http.StatusInternalServerError)
				return
			}
			if !valid {
				http.Error(w, "invalid or revoked api key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func validateWithCache(ctx context.Context, st *store.Store, c *cache.Cache, keyHash string) (bool, error) {
	cacheKey := "apikey:" + keyHash

	// checkng if it is cached
	cached, err := c.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		return cached == "valid", nil
	}

	// cache miss
	valid, err := st.ValidateAPIKey(ctx, keyHash)
	if err != nil {
		return false, err
	}

	// add cache for next time; cache both valid AND invlaid results -
	// this also protects the db from repeated lookups of a garbage/malicious key
	result := "invalid"
	if valid {
		result = "valid"
	}
	_ = c.Set(ctx, cacheKey, result, apiKeyCacheTTL) // ignore cache-write errors - caching is best-effor

	return valid, nil
}
