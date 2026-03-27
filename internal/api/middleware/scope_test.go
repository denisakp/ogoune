package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// setAuthContext injects auth_method and api_key_scope into the request context.
func setAuthContext(r *http.Request, authMethod string, scope domain.APIKeyScope) *http.Request {
	ctx := context.WithValue(r.Context(), "auth_method", authMethod)
	ctx = context.WithValue(ctx, "api_key_scope", scope)
	return r.WithContext(ctx)
}

// T028, T0M3 – RequireReadWrite: read-scoped API key is blocked on mutating routes.
func TestRequireReadWrite_ReadScopedAPIKey_Returns403(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/resources", nil)
	req = setAuthContext(req, "api_key", domain.APIKeyScopeRead)

	rec := httptest.NewRecorder()
	RequireReadWrite(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// T028 – RequireReadWrite: read_write-scoped API key is allowed.
func TestRequireReadWrite_ReadWriteScopedAPIKey_Passes(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/resources", nil)
	req = setAuthContext(req, "api_key", domain.APIKeyScopeReadWrite)

	rec := httptest.NewRecorder()
	RequireReadWrite(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// T028, T0M3 – RequireReadWrite: JWT session is never blocked.
func TestRequireReadWrite_JWTSession_Passes(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/resources", nil)
	req = setAuthContext(req, "jwt", domain.APIKeyScopeRead)

	rec := httptest.NewRecorder()
	RequireReadWrite(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// T028 – RequireReadWrite: unauthenticated request passes through (auth is handled by AuthMiddleware).
func TestRequireReadWrite_NoAuthMethod_Passes(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/resources", nil)

	rec := httptest.NewRecorder()
	RequireReadWrite(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// T028 – RequireJWTOnly: API key request is blocked.
func TestRequireJWTOnly_APIKeyAuth_Returns403(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/account/api-keys", nil)
	req = setAuthContext(req, "api_key", domain.APIKeyScopeReadWrite)

	rec := httptest.NewRecorder()
	RequireJWTOnly(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// T028 – RequireJWTOnly: JWT session is allowed.
func TestRequireJWTOnly_JWTAuth_Passes(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/account/api-keys", nil)
	req = setAuthContext(req, "jwt", domain.APIKeyScopeReadWrite)

	rec := httptest.NewRecorder()
	RequireJWTOnly(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// T028 – RequireJWTOnly: unauthenticated request passes through (caught later by AuthMiddleware).
func TestRequireJWTOnly_NoAuthMethod_Passes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/account/api-keys", nil)

	rec := httptest.NewRecorder()
	RequireJWTOnly(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// Error message content check for RequireReadWrite.
func TestRequireReadWrite_403ErrorMessage(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/resources/1", nil)
	req = setAuthContext(req, "api_key", domain.APIKeyScopeRead)

	rec := httptest.NewRecorder()
	RequireReadWrite(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "read-scoped")
}

// Error message content check for RequireJWTOnly.
func TestRequireJWTOnly_403ErrorMessage(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/account/api-keys/k1", nil)
	req = setAuthContext(req, "api_key", domain.APIKeyScopeReadWrite)

	rec := httptest.NewRecorder()
	RequireJWTOnly(okHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "session authentication")
}
