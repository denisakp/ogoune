package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/pkg/problemdetail"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(authService *service.AuthService, apiKeyService *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawAPIKey, isAPIKey := extractAPIKey(r)
			if isAPIKey {
				if apiKeyService == nil {
					problemdetail.Write(w, problemdetail.New("/problems/unauthorized", "Unauthorized", http.StatusUnauthorized, "unauthorized"))
					return
				}
				authenticated, err := apiKeyService.AuthenticateAPIKey(r.Context(), rawAPIKey)
				if err != nil {
					message := "invalid or revoked API key"
					typeURI := "/problems/key-revoked"
					if err == service.ErrAPIKeyExpired {
						message = "API key has expired"
						typeURI = "/problems/key-expired"
					}
					problemdetail.Write(w, problemdetail.New(typeURI, "Unauthorized", http.StatusUnauthorized, message))
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
						slog.Warn("failed to update api key usage", "key_id", keyID, "error", err)
					}
					slog.Info("api key authentication succeeded", "key_prefix", keyPrefix, "method", method, "path", path)
				}(authenticated.Key.ID, authenticated.Key.KeyPrefix, clientIP(r), r.Method, r.URL.Path)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if authService == nil {
				problemdetail.Write(w, problemdetail.New("/problems/unauthorized", "Unauthorized", http.StatusUnauthorized, "unauthorized"))
				return
			}

			token := extractToken(r)
			if token == "" {
				problemdetail.Write(w, problemdetail.New("/problems/unauthorized", "Unauthorized", http.StatusUnauthorized, "Missing authorization token"))
				return
			}

			// Validate token
			email, userID, err := authService.ValidateToken(token)
			if err != nil {
				problemdetail.Write(w, problemdetail.New("/problems/invalid-token", "Unauthorized", http.StatusUnauthorized, "Invalid or expired token"))
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
