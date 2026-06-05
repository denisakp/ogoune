package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func dummyHandler(body string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

func newRouterWith(domain, status string) (*HostRouter, http.Handler) {
	statusBundle := dummyHandler("STATUS_HTML", http.StatusOK)
	admin := dummyHandler("ADMIN_FALLBACK", http.StatusOK)
	hr := NewHostRouter(statusBundle)
	hr.Set(domain, status)
	return hr, hr.Middleware(admin)
}

func TestHostRouter_VerifiedMatch_ServesStatusBundle(t *testing.T) {
	_, h := newRouterWith("status.acme.com", "verified")
	req := httptest.NewRequest(http.MethodGet, "http://status.acme.com/", nil)
	req.Host = "status.acme.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "STATUS_HTML", rec.Body.String())
}

func TestHostRouter_Unverified_FallsThrough(t *testing.T) {
	_, h := newRouterWith("status.acme.com", "pending")
	req := httptest.NewRequest(http.MethodGet, "http://status.acme.com/", nil)
	req.Host = "status.acme.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, "ADMIN_FALLBACK", rec.Body.String())
}

func TestHostRouter_NormalizesPortSuffix(t *testing.T) {
	_, h := newRouterWith("status.acme.com", "verified")
	req := httptest.NewRequest(http.MethodGet, "http://status.acme.com:9596/", nil)
	req.Host = "status.acme.com:9596"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, "STATUS_HTML", rec.Body.String())
}

func TestHostRouter_NormalizesTrailingDot(t *testing.T) {
	_, h := newRouterWith("status.acme.com", "verified")
	req := httptest.NewRequest(http.MethodGet, "http://status.acme.com./", nil)
	req.Host = "status.acme.com."
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, "STATUS_HTML", rec.Body.String())
}

func TestHostRouter_AdminEndpoint_OnCustomHost_404(t *testing.T) {
	_, h := newRouterWith("status.acme.com", "verified")
	req := httptest.NewRequest(http.MethodGet, "http://status.acme.com/api/account/profile", nil)
	req.Host = "status.acme.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHostRouter_DefaultHost_PreservesAdminBehavior(t *testing.T) {
	_, h := newRouterWith("status.acme.com", "verified")
	req := httptest.NewRequest(http.MethodGet, "http://localhost/api/account/profile", nil)
	req.Host = "localhost"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ADMIN_FALLBACK", rec.Body.String())
}

func TestHostRouter_PublicAPI_OnCustomHost_Allowed(t *testing.T) {
	hr := NewHostRouter(dummyHandler("PUBLIC_STATUS_JSON", http.StatusOK))
	hr.Set("status.acme.com", "verified")
	admin := dummyHandler("admin", http.StatusOK)
	h := hr.Middleware(admin)

	for _, path := range []string{
		"/api/status", "/api/status/incidents", "/api/status/uptime",
		"/api/config/runtime", "/api/health",
	} {
		req := httptest.NewRequest(http.MethodGet, "http://status.acme.com"+path, nil)
		req.Host = "status.acme.com"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Equal(t, "PUBLIC_STATUS_JSON", rec.Body.String(), path)
	}
}

func TestHostRouter_NoConfig_AlwaysFallsThrough(t *testing.T) {
	hr := NewHostRouter(dummyHandler("status", http.StatusOK))
	// Never call Set — domain stays empty.
	admin := dummyHandler("admin", http.StatusOK)
	h := hr.Middleware(admin)
	req := httptest.NewRequest(http.MethodGet, "http://anywhere/whatever", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, "admin", rec.Body.String())
}
