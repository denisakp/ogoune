package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── mocks ────────────────────────────────────────────────────────────────────

type mockCredService struct {
	cred        *domain.ResourceCredential
	getErr      error
	setCreated  bool
	setErr      error
	deleteErr   error
	captured    *domain.ResourceCredential
	setCalls    int
	deleteCalls int
}

func (m *mockCredService) Get(_ context.Context, _ string) (*domain.ResourceCredential, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.cred, nil
}

func (m *mockCredService) Set(_ context.Context, resourceID, username string, password, options []byte) (bool, error) {
	m.setCalls++
	if m.setErr != nil {
		return false, m.setErr
	}
	m.captured = &domain.ResourceCredential{
		ResourceID: resourceID,
		Username:   username,
		Password:   password,
		Options:    options,
	}
	// Echo: subsequent Get returns the captured value with timestamps.
	m.cred = &domain.ResourceCredential{
		Base:       domain.Base{ID: "cred-1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		ResourceID: resourceID,
		Username:   username,
		Password:   password,
	}
	return m.setCreated, nil
}

func (m *mockCredService) Delete(_ context.Context, _ string) error {
	m.deleteCalls++
	return m.deleteErr
}

type mockTester struct {
	result domain.CheckResult
	err    error
}

func (m *mockTester) Test(_ context.Context, _ string, _ string, _, _ []byte) (domain.CheckResult, error) {
	return m.result, m.err
}

// ─── router helpers ───────────────────────────────────────────────────────────

func newCredentialRouter(svc *mockCredService, tester *mockTester) *chi.Mux {
	if tester == nil {
		tester = &mockTester{result: domain.CheckResult{Status: string(domain.StatusUp)}}
	}
	h := v1.NewResourceCredentialHandler(svc, tester)
	r := chi.NewRouter()
	r.Get("/api/v1/resources/{id}/credentials", h.Get)
	r.With(middleware.RequireReadWrite).Post("/api/v1/resources/{id}/credentials", h.Set)
	r.With(middleware.RequireReadWrite).Delete("/api/v1/resources/{id}/credentials", h.Delete)
	r.With(middleware.RequireReadWrite).Post("/api/v1/resources/{id}/credentials/test", h.Test)
	return r
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestCredentialHandler_Get_MaskedPassword(t *testing.T) {
	svc := &mockCredService{cred: &domain.ResourceCredential{
		Base:       domain.Base{ID: "c1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		ResourceID: "r1",
		Username:   "monitor",
		Password:   []byte("super-secret"),
	}}
	router := newCredentialRouter(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/r1/credentials", nil)
	req = injectReadScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.NotContains(t, body, "super-secret")
	assert.Contains(t, body, "••••••••")
}

func TestCredentialHandler_Get_NotFound(t *testing.T) {
	svc := &mockCredService{getErr: service.ErrCredentialNotFound}
	router := newCredentialRouter(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/r1/credentials", nil)
	req = injectReadScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "no credentials configured")
}

func TestCredentialHandler_Set_Created_Returns201(t *testing.T) {
	svc := &mockCredService{setCreated: true}
	router := newCredentialRouter(svc, nil)
	body := []byte(`{"username":"monitor","password":"s3cret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/r1/credentials", bytes.NewReader(body))
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	require.NotNil(t, svc.captured)
	assert.Equal(t, "monitor", svc.captured.Username)
	assert.Equal(t, "s3cret", string(svc.captured.Password))
}

func TestCredentialHandler_Set_Replaced_Returns200(t *testing.T) {
	svc := &mockCredService{setCreated: false}
	router := newCredentialRouter(svc, nil)
	body := []byte(`{"password":"new-pass"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/r1/credentials", bytes.NewReader(body))
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCredentialHandler_Set_MissingPassword_Returns422(t *testing.T) {
	svc := &mockCredService{}
	router := newCredentialRouter(svc, nil)
	body := []byte(`{"username":"monitor"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/r1/credentials", bytes.NewReader(body))
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Equal(t, 0, svc.setCalls)
}

func TestCredentialHandler_ScopeEnforcement_ReadKey_403OnWrites(t *testing.T) {
	router := newCredentialRouter(&mockCredService{}, nil)
	cases := []struct {
		method, path string
		body         []byte
	}{
		{"POST", "/api/v1/resources/r1/credentials", []byte(`{"password":"x"}`)},
		{"DELETE", "/api/v1/resources/r1/credentials", nil},
		{"POST", "/api/v1/resources/r1/credentials/test", []byte(`{"password":"x"}`)},
	}
	for _, c := range cases {
		t.Run(c.method+" "+c.path, func(t *testing.T) {
			var body *bytes.Reader
			if c.body != nil {
				body = bytes.NewReader(c.body)
			} else {
				body = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(c.method, c.path, body)
			req = injectReadScope(req)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusForbidden, rr.Code)
		})
	}
}

func TestCredentialHandler_Delete_Success_Returns204(t *testing.T) {
	svc := &mockCredService{}
	router := newCredentialRouter(svc, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/resources/r1/credentials", nil)
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, 1, svc.deleteCalls)
}

func TestCredentialHandler_Delete_NotFound_Returns404(t *testing.T) {
	svc := &mockCredService{deleteErr: service.ErrCredentialNotFound}
	router := newCredentialRouter(svc, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/resources/r1/credentials", nil)
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCredentialHandler_Test_Success(t *testing.T) {
	svc := &mockCredService{}
	tester := &mockTester{result: domain.CheckResult{Status: string(domain.StatusUp)}}
	router := newCredentialRouter(svc, tester)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/r1/credentials/test",
		bytes.NewReader([]byte(`{"password":"s3cret"}`)))
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "ok", data["status"])
}

func TestCredentialHandler_Test_AuthFailedCause(t *testing.T) {
	svc := &mockCredService{}
	cause := domain.ProtocolAuthFailed
	tester := &mockTester{result: domain.CheckResult{
		Status: string(domain.StatusDown),
		Cause:  &cause,
	}}
	router := newCredentialRouter(svc, tester)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/r1/credentials/test",
		bytes.NewReader([]byte(`{"password":"s3cret"}`)))
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "failed", data["status"])
	assert.Equal(t, "auth_failed", data["cause"])
}

func TestCredentialHandler_Test_ResourceNotFound(t *testing.T) {
	tester := &mockTester{err: service.ErrResourceNotFound}
	router := newCredentialRouter(&mockCredService{}, tester)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/r1/credentials/test",
		bytes.NewReader([]byte(`{"password":"x"}`)))
	req = injectReadWriteScope(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ─── Audit log emission (T035) ────────────────────────────────────────────────

func TestCredentialHandler_AuditLog_OnCreate(t *testing.T) {
	captured, restore := captureSlog(t)
	defer restore()

	svc := &mockCredService{setCreated: true}
	router := newCredentialRouter(svc, nil)
	body := []byte(`{"username":"monitor","password":"super-secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/r1/credentials", bytes.NewReader(body))
	req = injectReadWriteScopeWithUser(req, "user-42")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	got := captured.String()
	assert.Contains(t, got, "credential.create")
	assert.Contains(t, got, "user-42")
	assert.Contains(t, got, "r1")
	assert.NotContains(t, got, "super-secret", "audit log must never include the password")
}

func TestCredentialHandler_AuditLog_OnDelete(t *testing.T) {
	captured, restore := captureSlog(t)
	defer restore()

	svc := &mockCredService{}
	router := newCredentialRouter(svc, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/resources/r1/credentials", nil)
	req = injectReadWriteScopeWithUser(req, "user-42")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)

	got := captured.String()
	assert.Contains(t, got, "credential.delete")
	assert.Contains(t, got, "user-42")
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func injectReadWriteScopeWithUser(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), "auth_method", "api_key")
	ctx = context.WithValue(ctx, "api_key_scope", domain.APIKeyScopeReadWrite)
	ctx = context.WithValue(ctx, "user_id", userID)
	return r.WithContext(ctx)
}

func captureSlog(t *testing.T) (*strings.Builder, func()) {
	t.Helper()
	var buf strings.Builder
	prev := slog.Default()
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))
	return &buf, func() { slog.SetDefault(prev) }
}

// Ensure errors import is used even when no error cases reference it directly.
var _ = errors.New
