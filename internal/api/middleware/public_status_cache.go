package middleware

import (
	"net/http"
	"strconv"
)

// PublicStatusCache adds short-lived HTTP cache headers for the public status
// page JSON endpoints (spec 060 SC-006). It lets reverse proxies and browsers
// serve cached responses for `maxAgeSeconds` and revalidate in the background
// for an additional `staleSeconds` window.
//
// Default tuning: max-age=15, stale-while-revalidate=30. Per FR-026 the
// public payload must not be older than 60s, which the upstream cron + cache
// chain enforces; this middleware only governs HTTP-layer caching.
func PublicStatusCache(maxAgeSeconds, staleSeconds int) func(http.Handler) http.Handler {
	if maxAgeSeconds <= 0 {
		maxAgeSeconds = 15
	}
	if staleSeconds < 0 {
		staleSeconds = 0
	}
	cc := "public, max-age=" + strconv.Itoa(maxAgeSeconds) +
		", stale-while-revalidate=" + strconv.Itoa(staleSeconds)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", cc)
			w.Header().Set("Vary", "Accept-Encoding")
			next.ServeHTTP(w, r)
		})
	}
}
