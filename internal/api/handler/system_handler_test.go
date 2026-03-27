package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSystemHandlerGetEditionCommunity(t *testing.T) {
	os.Unsetenv("ENTERPRISE_LICENSE_KEY")
	os.Unsetenv("APP_VERSION")

	h := NewSystemHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/system/edition", nil)
	rr := httptest.NewRecorder()

	h.GetEdition(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["edition"] != "community" {
		t.Fatalf("expected edition community, got %q", body["edition"])
	}
	if body["version"] != "1.0.0" {
		t.Fatalf("expected default version 1.0.0, got %q", body["version"])
	}
}

func TestSystemHandlerGetEditionEnterprise(t *testing.T) {
	os.Setenv("ENTERPRISE_LICENSE_KEY", "pg_ent_abc123")
	os.Setenv("APP_VERSION", "1.2.3")
	defer os.Unsetenv("ENTERPRISE_LICENSE_KEY")
	defer os.Unsetenv("APP_VERSION")

	h := NewSystemHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/system/edition", nil)
	rr := httptest.NewRecorder()

	h.GetEdition(rr, req)

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["edition"] != "enterprise" {
		t.Fatalf("expected edition enterprise, got %q", body["edition"])
	}
	if body["version"] != "1.2.3" {
		t.Fatalf("expected version 1.2.3, got %q", body["version"])
	}
}
