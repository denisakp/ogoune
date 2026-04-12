package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func newTestRegistry(t *testing.T) *prometheus.Registry {
	t.Helper()
	return prometheus.NewRegistry()
}

// T017: Bearer token middleware — four cases.

// Case 1: token configured + correct header → 200
func TestNewHandler_TokenCorrect(t *testing.T) {
	reg := newTestRegistry(t)
	h := NewHandler("test-secret", reg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer test-secret")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Case 2: token configured + wrong token → 401
func TestNewHandler_TokenWrong(t *testing.T) {
	reg := newTestRegistry(t)
	h := NewHandler("test-secret", reg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Case 3: token configured + absent header → 401
func TestNewHandler_TokenAbsent(t *testing.T) {
	reg := newTestRegistry(t)
	h := NewHandler("test-secret", reg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Case 4: no token configured + no header → 200 (unauthenticated open access)
func TestNewHandler_NoToken(t *testing.T) {
	reg := newTestRegistry(t)
	h := NewHandler("", reg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
