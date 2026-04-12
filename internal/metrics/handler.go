package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHandler returns an http.Handler for the /metrics endpoint.
// When token is empty the base promhttp handler is returned directly (unauthenticated).
// When token is set, requests must supply "Authorization: Bearer <token>" or receive 401.
func NewHandler(token string, reg prometheus.Gatherer) http.Handler {
	base := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	if token == "" {
		return base
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		base.ServeHTTP(w, r)
	})
}
