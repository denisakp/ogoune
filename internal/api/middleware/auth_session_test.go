package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Spec 059 FR-009 — revoking a session must take effect on the very next
// request, with no cache layer between the middleware and the repository.
func TestAuthMiddleware_RevokedSession_NextRequestReturns401(t *testing.T) {
	userRepo := fake.NewUserRepository()
	apiKeyRepo := fake.NewAPIKeyRepository()
	sessionRepo := fake.NewSessionRepository()

	jwtMgr := service.NewJWTManager("test-secret", "ogoune", time.Hour)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, userRepo)
	sessionSvc := service.NewSessionService(sessionRepo)
	authSvc.SetSessionService(sessionSvc)

	sess, err := sessionSvc.Issue(t.Context(), "u1", "Chrome/138", "1.2.3.4")
	require.NoError(t, err)

	token, err := jwtMgr.GenerateWithSession(t.Context(), "u1@x.test", "u1", sess.ID)
	require.NoError(t, err)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := AuthMiddleware(authSvc, apiKeySvc, sessionSvc)(next)

	// First request: 200.
	req1 := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	rec1 := httptest.NewRecorder()
	mw.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code, "fresh session must pass")

	// Revoke immediately (no sleep).
	require.NoError(t, sessionRepo.Revoke(t.Context(), sess.ID, time.Now()))

	// Second request: 401 SESSION_REVOKED.
	req2 := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()
	mw.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "session-revoked")
}

// A token without a sid (legacy JWT predating spec 059) MUST pass through the
// middleware untouched even when the session service is wired.
func TestAuthMiddleware_TokenWithoutSID_PassesThrough(t *testing.T) {
	userRepo := fake.NewUserRepository()
	apiKeyRepo := fake.NewAPIKeyRepository()
	sessionRepo := fake.NewSessionRepository()
	jwtMgr := service.NewJWTManager("test-secret", "ogoune", time.Hour)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, userRepo)
	sessionSvc := service.NewSessionService(sessionRepo)

	token, err := jwtMgr.Generate(t.Context(), "u1@x.test", "u1")
	require.NoError(t, err)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := AuthMiddleware(authSvc, apiKeySvc, sessionSvc)(next)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
