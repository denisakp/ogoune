package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/denisakp/ogoune/internal/config"
)

func TestRuntimeConfigHandler_Get(t *testing.T) {
	cfg := &config.Config{SSLProvider: "external"}
	h := NewRuntimeConfigHandler(cfg, "0.4.2-test")

	req := httptest.NewRequest(http.MethodGet, "/api/config/runtime", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	var got runtimeConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.SSLProvider != "external" {
		t.Errorf("ssl_provider: got %q want external", got.SSLProvider)
	}
	if got.Edition != "community" && got.Edition != "enterprise" {
		t.Errorf("edition: got %q want community|enterprise", got.Edition)
	}
	if got.Version != "0.4.2-test" {
		t.Errorf("version: got %q want 0.4.2-test", got.Version)
	}
	// Community by default (no ENTERPRISE_LICENSE_KEY in test env) → powered_by_required must be true.
	if !got.PoweredByRequired {
		t.Errorf("powered_by_required: got false want true in community mode")
	}
}

func TestRuntimeConfigHandler_PoweredByRequired_EnterpriseSuppresses(t *testing.T) {
	t.Setenv("ENTERPRISE_LICENSE_KEY", "pg_ent_test_key")
	cfg := &config.Config{SSLProvider: "external"}
	h := NewRuntimeConfigHandler(cfg, "x")
	rec := httptest.NewRecorder()
	h.Get(rec, httptest.NewRequest(http.MethodGet, "/api/config/runtime", nil))
	var got runtimeConfigResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got.Edition != "enterprise" {
		t.Fatalf("edition: got %q want enterprise", got.Edition)
	}
	if got.PoweredByRequired {
		t.Errorf("powered_by_required: got true want false in enterprise mode")
	}
}

func TestRuntimeConfigHandler_AllSSLProviders(t *testing.T) {
	for _, mode := range []string{"letsencrypt", "external", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			cfg := &config.Config{SSLProvider: mode}
			h := NewRuntimeConfigHandler(cfg, "x")
			rec := httptest.NewRecorder()
			h.Get(rec, httptest.NewRequest(http.MethodGet, "/api/config/runtime", nil))
			var got runtimeConfigResponse
			_ = json.Unmarshal(rec.Body.Bytes(), &got)
			if got.SSLProvider != mode {
				t.Errorf("ssl_provider: got %q want %q", got.SSLProvider, mode)
			}
		})
	}
}
