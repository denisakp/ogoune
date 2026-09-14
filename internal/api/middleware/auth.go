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

const problemUnauthorized = "/problems/unauthorized"

// AuthMiddleware creates a middleware that validates JWT tokens.
// When the JWT carries a sid claim, the middleware also consults
// sessions.revoked_at and refuses the request if the session has been revoked
func AuthMiddleware(authService *service.AuthService, apiKeyService *service.APIKeyService, sessionService *service.SessionService) func(http.Handler) http.Handler {
	a := &authenticator{auth: authService, apiKeys: apiKeyService, sessions: sessionService}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// An API key, when presented, is the only credential considered;
			// a JWT is never consulted as a fallback for a bad key.
			var ctx context.Context
			var ok bool
			if rawAPIKey, isAPIKey := extractAPIKey(r); isAPIKey {
				ctx, ok = a.withAPIKey(w, r, rawAPIKey)
			} else {
				ctx, ok = a.withJWT(w, r)
			}
			if !ok {
				return // the problem response has been written
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authenticator holds the three services the middleware may consult. Each
// method writes the problem response itself on failure and reports false, so
// the middleware body is one decision: which credential, then next.
type authenticator struct {
	auth     *service.AuthService
	apiKeys  *service.APIKeyService
	sessions *service.SessionService
}

func unauthorized(w http.ResponseWriter, typeURI, detail string) {
	problemdetail.Write(w, problemdetail.New(typeURI, "Unauthorized", http.StatusUnauthorized, detail))
}

// withAPIKey authenticates an API key and returns a context carrying the
// key's identity and scope. The last-used bump and the audit line run in the
// background so they never add to request latency.
func (a *authenticator) withAPIKey(w http.ResponseWriter, r *http.Request, rawAPIKey string) (context.Context, bool) {
	if a.apiKeys == nil {
		unauthorized(w, problemUnauthorized, "unauthorized")
		return nil, false
	}
	authenticated, err := a.apiKeys.AuthenticateAPIKey(r.Context(), rawAPIKey)
	if err != nil {
		message := "invalid or revoked API key"
		typeURI := "/problems/key-revoked"
		if err == service.ErrAPIKeyExpired {
			message = "API key has expired"
			typeURI = "/problems/key-expired"
		}
		unauthorized(w, typeURI, message)
		return nil, false
	}

	ctx := context.WithValue(r.Context(), "email", authenticated.User.Email)
	ctx = context.WithValue(ctx, "user_id", authenticated.User.ID)
	ctx = context.WithValue(ctx, "auth_method", "api_key")
	ctx = context.WithValue(ctx, "api_key_scope", authenticated.Key.Scope)
	ctx = context.WithValue(ctx, "api_key_id", authenticated.Key.ID)

	go func(keyID, keyPrefix, ip, method, path string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := a.apiKeys.UpdateLastUsed(bgCtx, keyID, ip); err != nil {
			slog.Warn("failed to update api key usage", "key_id", keyID, "error", err)
		}
		slog.Info("api key authentication succeeded", "key_prefix", keyPrefix, "method", method, "path", path)
	}(authenticated.Key.ID, authenticated.Key.KeyPrefix, clientIP(r), r.Method, r.URL.Path)

	return ctx, true
}

// withJWT validates the bearer token, refuses a revoked session when the
// token names one, and returns a context carrying the user's identity with
// read-write scope. The session's last-active bump runs in the background.
func (a *authenticator) withJWT(w http.ResponseWriter, r *http.Request) (context.Context, bool) {
	if a.auth == nil {
		unauthorized(w, problemUnauthorized, "unauthorized")
		return nil, false
	}
	token := extractToken(r)
	if token == "" {
		unauthorized(w, problemUnauthorized, "Missing authorization token")
		return nil, false
	}

	// Validate token (including optional sid claim).
	email, userID, sessionID, err := a.auth.ValidateTokenFull(token)
	if err != nil {
		unauthorized(w, "/problems/invalid-token", "Invalid or expired token")
		return nil, false
	}

	// Refuse revoked sessions immediately.
	if a.sessions != nil && sessionID != "" {
		if err := a.sessions.Validate(r.Context(), sessionID); err != nil {
			unauthorized(w, "/problems/session-revoked", "Session no longer valid")
			return nil, false
		}
		// Fire-and-forget last-active bump (best-effort).
		go func(id string) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := a.sessions.TouchLastActive(bgCtx, id); err != nil {
				slog.Warn("failed to touch session last_active_at", "session_id", id, "error", err)
			}
		}(sessionID)
	}

	ctx := context.WithValue(r.Context(), "email", email)
	ctx = context.WithValue(ctx, "user_id", userID)
	ctx = context.WithValue(ctx, "session_id", sessionID)
	ctx = context.WithValue(ctx, "auth_method", "jwt")
	ctx = context.WithValue(ctx, "api_key_scope", domain.APIKeyScopeReadWrite)
	return ctx, true
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
