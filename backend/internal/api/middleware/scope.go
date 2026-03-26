package middleware

import (
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/domain"
)

// RequireReadWrite ensures read-scoped API keys cannot access mutating routes.
func RequireReadWrite(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authMethod, _ := r.Context().Value("auth_method").(string)
		if authMethod != "api_key" {
			next.ServeHTTP(w, r)
			return
		}

		scope, _ := r.Context().Value("api_key_scope").(domain.APIKeyScope)
		if scope == domain.APIKeyScopeRead {
			response.Error(w, http.StatusForbidden, "read-scoped API keys cannot perform write operations")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireJWTOnly blocks API-key-authenticated requests from protected management endpoints.
func RequireJWTOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authMethod, _ := r.Context().Value("auth_method").(string)
		if authMethod == "api_key" {
			response.Error(w, http.StatusForbidden, "API key management requires session authentication")
			return
		}
		next.ServeHTTP(w, r)
	})
}
