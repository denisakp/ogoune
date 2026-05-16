package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeadersConfig holds configuration for the security headers middleware.
type SecurityHeadersConfig struct {
	AppEnv            string
	EnableSwagger     bool
	SwaggerPathPrefix string
}

// defaultCSP is the strict Content-Security-Policy applied to all routes
// except Swagger UI docs when enabled.
const defaultCSP = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'"

// swaggerCSP is a relaxed Content-Security-Policy for Swagger UI routes,
// allowing assets from the Swagger CDN.
const swaggerCSP = "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://unpkg.com; img-src 'self' data: https://validator.swagger.io; frame-ancestors 'none'"

// SecurityHeaders returns middleware that sets standard security headers on
// every response. HSTS is conditional on production environment or TLS.
// CSP is relaxed for Swagger UI routes when enabled.
func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Static security headers (always set)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			// CSP: relaxed for Swagger UI docs, strict everywhere else
			csp := defaultCSP
			if cfg.EnableSwagger && cfg.SwaggerPathPrefix != "" && strings.HasPrefix(r.URL.Path, cfg.SwaggerPathPrefix) {
				csp = swaggerCSP
			}
			w.Header().Set("Content-Security-Policy", csp)

			// HSTS: only in production or when TLS is detected
			if cfg.AppEnv == "production" || isTLS(r) {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isTLS returns true if the request arrived over a secure connection,
// either directly (r.TLS) or via a reverse proxy (X-Forwarded-Proto, Forwarded).
func isTLS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	forwarded := r.Header.Get("Forwarded")
	if forwarded != "" && strings.Contains(strings.ToLower(forwarded), "proto=https") {
		return true
	}
	return false
}
