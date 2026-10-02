package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/pkg/crypto"
)

// Spec 095, SC-004: once a former account is erased, its sessions and its API
// keys are refused by the real authentication middleware on the very next
// request; the operator's own session keeps working.
func TestErasedAccount_IsRefusedByTheAuthMiddleware(t *testing.T) {
	t.Setenv("APP_SECRET_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	ctx := context.Background()
	rt := internaltest.SetupSQLite(t).Runtime

	users := store.NewUserRepositorySQLC(rt)
	sessions := store.NewSessionRepositorySQLC(rt)
	keys := store.NewAPIKeyRepositorySQLC(rt)
	for _, u := range []*domain.User{
		{Base: domain.Base{ID: "u-op"}, Email: "op@example.com", Name: "Operator", HashedPassword: "x"},
		{Base: domain.Base{ID: "u-old"}, Email: "old@example.com", Name: "Old admin", HashedPassword: "x"},
	} {
		_, err := users.Create(ctx, u)
		require.NoError(t, err)
	}

	jwtMgr := service.NewJWTManager("test-secret-key-at-least-32-bytes-long", "ogoune", time.Hour)
	authSvc := service.NewAuthService(users, jwtMgr)
	apiKeySvc := service.NewAPIKeyService(keys, users)
	sessionSvc := service.NewSessionService(sessions)
	authSvc.SetSessionService(sessionSvc)

	tokenFor := func(userID, email string) string {
		s, err := sessionSvc.Issue(ctx, userID, "Firefox", "203.0.113.9")
		require.NoError(t, err)
		tok, err := jwtMgr.GenerateWithSession(ctx, email, userID, s.ID)
		require.NoError(t, err)
		return tok
	}
	opToken, oldToken := tokenFor("u-op", "op@example.com"), tokenFor("u-old", "old@example.com")
	oldKey, err := apiKeySvc.CreateAPIKey(ctx, "u-old", "ci", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)

	mw := middleware.AuthMiddleware(authSvc, apiKeySvc, sessionSvc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	call := func(header, value string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors", nil)
		req.Header.Set(header, value)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		return rec.Code
	}
	require.Equal(t, http.StatusOK, call("Authorization", "Bearer "+oldToken))
	require.Equal(t, http.StatusOK, call("X-API-Key", oldKey.Key))

	eraser := service.NewErasureService(users, sessions, keys, store.NewIncidentUpdateRepositorySQLC(rt),
		store.NewNotificationChannelRepositorySQLC(rt), store.NewReportSettingsRepositorySQLC(rt),
		store.NewReportHistoryRepositorySQLC(rt), store.NewErasureRepositorySQLC(rt))
	_, err = eraser.Erase(ctx, "u-op", service.ErasureSubject{AccountID: "u-old"}, "old@example.com",
		func(context.Context) error { return nil })
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, call("Authorization", "Bearer "+oldToken), "the erased account's session is refused")
	assert.Equal(t, http.StatusUnauthorized, call("X-API-Key", oldKey.Key), "its API key is refused")
	assert.Equal(t, http.StatusOK, call("Authorization", "Bearer "+opToken), "the operator stays signed in")
}
