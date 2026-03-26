package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/service"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(authService *service.AuthService, apiKeyService *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawAPIKey, isAPIKey := extractAPIKey(r)
			if isAPIKey {
				if apiKeyService == nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
					return
				}
				authenticated, err := apiKeyService.AuthenticateAPIKey(r.Context(), rawAPIKey)
				if err != nil {
					status := http.StatusUnauthorized
					message := "invalid or revoked API key"
					if err == service.ErrAPIKeyExpired {
						message = "API key has expired"
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error": message,
					})
					return
				}

				ctx := context.WithValue(r.Context(), "email", authenticated.User.Email)
				ctx = context.WithValue(ctx, "user_id", authenticated.User.ID)
				ctx = context.WithValue(ctx, "auth_method", "api_key")
				ctx = context.WithValue(ctx, "api_key_scope", authenticated.Key.Scope)
				ctx = context.WithValue(ctx, "api_key_id", authenticated.Key.ID)

				go func(keyID, keyPrefix, ip, method, path string) {
					bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					if err := apiKeyService.UpdateLastUsed(bgCtx, keyID, ip); err != nil {
						log.Printf("[WARN] failed to update api key usage for key_id=%s: %v", keyID, err)
					}
					log.Printf("api_key_auth_success key_prefix=%s method=%s path=%s", keyPrefix, method, path)
				}(authenticated.Key.ID, authenticated.Key.KeyPrefix, clientIP(r), r.Method, r.URL.Path)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if authService == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}

			token := extractToken(r)
			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "unauthorized",
					"message": "Missing authorization token",
				})
				return
			}

			// Validate token
			email, userID, err := authService.ValidateToken(token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "unauthorized",
					"message": "Invalid or expired token",
				})
				return
			}

			// Add email and userID to request context
			ctx := context.WithValue(r.Context(), "email", email)
			ctx = context.WithValue(ctx, "user_id", userID)
			ctx = context.WithValue(ctx, "auth_method", "jwt")
			ctx = context.WithValue(ctx, "api_key_scope", domain.APIKeyScopeReadWrite)

			// Token is valid, proceed to next handler with enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractAPIKey(r *http.Request) (string, bool) {
	if key := strings.TrimSpace(r.Header.Get("X-API-Key")); key != "" {
		return key, true
	}

	bearerToken := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(bearerToken, "Bearer pk_live_") {
		return strings.TrimSpace(bearerToken[7:]), true
	}

	return "", false
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	return strings.TrimSpace(r.RemoteAddr)
}

// extractToken extracts JWT token from Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		return bearerToken[7:]
	}
	return ""
}
