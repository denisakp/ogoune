package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAPIKeyService is a test double used by account handler tests.
type mockAPIKeyService struct {
	createFunc func(ctx context.Context, userID, name string, scope domain.APIKeyScope, expiresAt *time.Time) (*dto.CreateAPIKeyResponse, error)
	listFunc   func(ctx context.Context, userID string) ([]dto.APIKeyListItem, error)
	revokeFunc func(ctx context.Context, userID, keyID string) error
}

func (m *mockAPIKeyService) CreateAPIKey(ctx context.Context, userID, name string, scope domain.APIKeyScope, expiresAt *time.Time) (*dto.CreateAPIKeyResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, name, scope, expiresAt)
	}
	return &dto.CreateAPIKeyResponse{
		ID:        "key-1",
		Name:      name,
		Key:       "pk_live_testkey",
		KeyPrefix: "pk_live_tes",
		Scope:     scope,
	}, nil
}

func (m *mockAPIKeyService) ListAPIKeys(ctx context.Context, userID string) ([]dto.APIKeyListItem, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userID)
	}
	return []dto.APIKeyListItem{}, nil
}

func (m *mockAPIKeyService) RevokeAPIKey(ctx context.Context, userID, keyID string) error {
	if m.revokeFunc != nil {
		return m.revokeFunc(ctx, userID, keyID)
	}
	return nil
}

// newAccountHandlerForTest creates an AccountHandler wired to mock services with no auth service needed.
func newAccountHandlerForTest(apiKeySvc APIKeyService) *AccountHandler {
	return &AccountHandler{apiKeyService: apiKeySvc}
}

// withUserID injects user_id into the request context (simulating passed auth middleware).
func withUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), "user_id", userID)
	return r.WithContext(ctx)
}

// T015 – POST /account/api-keys returns 201 with one-time key reveal.
func TestAccountHandler_CreateAPIKey_Success(t *testing.T) {
	h := newAccountHandlerForTest(&mockAPIKeyService{})

	body := `{"name":"CI Pipeline","scope":"read_write"}`
	req := httptest.NewRequest(http.MethodPost, "/account/api-keys", bytes.NewBufferString(body))
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()

	h.CreateAPIKey(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp dto.CreateAPIKeyResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Key)
	assert.NotEmpty(t, resp.ID)
}

func TestAccountHandler_CreateAPIKey_InvalidJSON(t *testing.T) {
	h := newAccountHandlerForTest(&mockAPIKeyService{})

	req := httptest.NewRequest(http.MethodPost, "/account/api-keys", bytes.NewBufferString("{invalid}"))
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()

	h.CreateAPIKey(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAccountHandler_CreateAPIKey_ValidationFailed(t *testing.T) {
	svc := &mockAPIKeyService{
		createFunc: func(_ context.Context, _, _ string, _ domain.APIKeyScope, _ *time.Time) (*dto.CreateAPIKeyResponse, error) {
			return nil, service.ErrValidationFailed
		},
	}
	h := newAccountHandlerForTest(svc)

	body := `{"name":"","scope":"read"}`
	req := httptest.NewRequest(http.MethodPost, "/account/api-keys", bytes.NewBufferString(body))
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()

	h.CreateAPIKey(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAccountHandler_CreateAPIKey_LimitReached(t *testing.T) {
	svc := &mockAPIKeyService{
		createFunc: func(_ context.Context, _, _ string, _ domain.APIKeyScope, _ *time.Time) (*dto.CreateAPIKeyResponse, error) {
			return nil, service.ErrAPIKeyLimitReached
		},
	}
	h := newAccountHandlerForTest(svc)

	body := `{"name":"Key","scope":"read"}`
	req := httptest.NewRequest(http.MethodPost, "/account/api-keys", bytes.NewBufferString(body))
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()

	h.CreateAPIKey(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// T022 – GET /account/api-keys lists all keys.
func TestAccountHandler_ListAPIKeys_Success(t *testing.T) {
	now := time.Now()
	svc := &mockAPIKeyService{
		listFunc: func(_ context.Context, _ string) ([]dto.APIKeyListItem, error) {
			return []dto.APIKeyListItem{
				{ID: "k1", Name: "CI", KeyPrefix: "pk_live_ci", Scope: domain.APIKeyScopeReadWrite, IsActive: true, CreatedAt: now},
				{ID: "k2", Name: "Monitoring", KeyPrefix: "pk_live_mo", Scope: domain.APIKeyScopeRead, IsActive: false, CreatedAt: now},
			}, nil
		},
	}
	h := newAccountHandlerForTest(svc)

	req := httptest.NewRequest(http.MethodGet, "/account/api-keys", nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()

	h.ListAPIKeys(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var items []dto.APIKeyListItem
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&items))
	assert.Len(t, items, 2)
	// Raw key field must not be present in list response payloads.
	raw, _ := json.Marshal(items)
	assert.NotContains(t, string(raw), "\"key\":")
}

func TestAccountHandler_ListAPIKeys_ServiceError(t *testing.T) {
	svc := &mockAPIKeyService{
		listFunc: func(_ context.Context, _ string) ([]dto.APIKeyListItem, error) {
			return nil, errors.New("db error")
		},
	}
	h := newAccountHandlerForTest(svc)

	req := httptest.NewRequest(http.MethodGet, "/account/api-keys", nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()

	h.ListAPIKeys(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// T022 – DELETE /account/api-keys/{id} revokes a key.
func TestAccountHandler_RevokeAPIKey_Success(t *testing.T) {
	revoked := ""
	svc := &mockAPIKeyService{
		revokeFunc: func(_ context.Context, _, keyID string) error {
			revoked = keyID
			return nil
		},
	}
	h := newAccountHandlerForTest(svc)

	req := httptest.NewRequest(http.MethodDelete, "/account/api-keys/key-abc", nil)
	req = withUserID(req, "user-1")
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "key-abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	rec := httptest.NewRecorder()

	h.RevokeAPIKey(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "key-abc", revoked)
}

func TestAccountHandler_RevokeAPIKey_NotFound(t *testing.T) {
	svc := &mockAPIKeyService{
		revokeFunc: func(_ context.Context, _, _ string) error {
			return service.ErrAPIKeyNotFound
		},
	}
	h := newAccountHandlerForTest(svc)

	req := httptest.NewRequest(http.MethodDelete, "/account/api-keys/missing", nil)
	req = withUserID(req, "user-1")
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "missing")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	rec := httptest.NewRecorder()

	h.RevokeAPIKey(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAccountHandler_RevokeAPIKey_MissingID(t *testing.T) {
	h := newAccountHandlerForTest(&mockAPIKeyService{})

	req := httptest.NewRequest(http.MethodDelete, "/account/api-keys/", nil)
	req = withUserID(req, "user-1")
	chiCtx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	rec := httptest.NewRecorder()

	h.RevokeAPIKey(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
