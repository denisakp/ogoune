package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/denisakp/pulseguard/internal/service"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			token := extractToken(r)
			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "unauthorized",
					"message": "Missing authorization token",
				})
				return
			}

			// Validate token
			_, err := authService.ValidateToken(token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "unauthorized",
					"message": "Invalid or expired token",
				})
				return
			}

			// Token is valid, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// extractToken extracts JWT token from Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		return bearerToken[7:]
	}
	return ""
}
