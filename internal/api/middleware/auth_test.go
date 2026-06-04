package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildAuthMiddleware(t *testing.T) (http.Handler, *service.APIKeyService, *fake.UserRepository, *service.AuthService) {
	t.Helper()

	userRepo := fake.NewUserRepository()
	apiKeyRepo := fake.NewAPIKeyRepository()
	jwtManager := service.NewJWTManager("test-secret", "ogoune", 24*time.Hour)
	authSvc := service.NewAuthService(userRepo, jwtManager)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, userRepo)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(authSvc, apiKeySvc, nil)(next)
	return handler, apiKeySvc, userRepo, authSvc
}

func seedUserForAuth(t *testing.T, userRepo *fake.UserRepository) *domain.User {
	t.Helper()

	user := &domain.User{
		Email:               "auth@example.com",
		HashedPassword:      "hash",
		PasswordInitialized: true,
	}
	created, err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)
	return created
}

func TestAuthMiddleware_XAPIKey_ValidKey_Accepted(t *testing.T) {
	handler, apiKeySvc, userRepo, _ := buildAuthMiddleware(t)
	user := seedUserForAuth(t, userRepo)

	created, err := apiKeySvc.CreateAPIKey(context.Background(), user.ID, "CI", domain.APIKeyScopeReadWrite, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req.Header.Set("X-API-Key", created.Key)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_BearerAPIKey_ValidKey_Accepted(t *testing.T) {
	handler, apiKeySvc, userRepo, _ := buildAuthMiddleware(t)
	user := seedUserForAuth(t, userRepo)

	created, err := apiKeySvc.CreateAPIKey(context.Background(), user.ID, "Deploy", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req.Header.Set("Authorization", "Bearer "+created.Key)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_APIKey_SetsContextValues(t *testing.T) {
	userRepo := fake.NewUserRepository()
	apiKeyRepo := fake.NewAPIKeyRepository()
	jwtMgr := service.NewJWTManager("secret", "ogoune", time.Hour)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, userRepo)

	user := &domain.User{Email: "ctx@example.com", HashedPassword: "h", PasswordInitialized: true}
	created, err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)
	key, err := apiKeySvc.CreateAPIKey(context.Background(), created.ID, "ctx-key", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)

	var capturedCtx context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})
	handler := AuthMiddleware(authSvc, apiKeySvc, nil)(next)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req.Header.Set("X-API-Key", key.Key)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "api_key", capturedCtx.Value("auth_method"))
	assert.Equal(t, created.ID, capturedCtx.Value("user_id"))
	assert.Equal(t, domain.APIKeyScopeRead, capturedCtx.Value("api_key_scope"))
}

func TestAuthMiddleware_RawKeyNotExposedInErrorResponse(t *testing.T) {
	handler, _, _, _ := buildAuthMiddleware(t)
	rawKey := "pk_live_0000000000000000000000000000000000000000"

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req.Header.Set("X-API-Key", rawKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NotContains(t, rec.Body.String(), rawKey)
}

func TestAuthMiddleware_RevokedKey_Returns401(t *testing.T) {
	handler, apiKeySvc, userRepo, _ := buildAuthMiddleware(t)
	user := seedUserForAuth(t, userRepo)
	created, err := apiKeySvc.CreateAPIKey(context.Background(), user.ID, "Temp", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)
	err = apiKeySvc.RevokeAPIKey(context.Background(), user.ID, created.ID)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	req.Header.Set("X-API-Key", created.Key)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ExpiredKey_Returns401(t *testing.T) {
	userRepo := fake.NewUserRepository()
	apiKeyRepo := fake.NewAPIKeyRepository()
	jwtMgr := service.NewJWTManager("secret", "ogoune", time.Hour)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, userRepo)

	user := &domain.User{Email: "expire@example.com", HashedPassword: "h", PasswordInitialized: true}
	created, err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)
	future := time.Now().Add(time.Hour)
	key, err := apiKeySvc.CreateAPIKey(context.Background(), created.ID, "Expiring", domain.APIKeyScopeRead, &future)
	require.NoError(t, err)

	apiKeySvc.SetNow(func() time.Time { return time.Now().Add(2 * time.Hour) })
	mw := AuthMiddleware(authSvc, apiKeySvc, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", key.Key)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "expired")
}

func TestAuthMiddleware_NoToken_Returns401(t *testing.T) {
	handler, _, _, _ := buildAuthMiddleware(t)
	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
