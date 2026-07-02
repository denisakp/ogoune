package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestServeOpenAPISpec_FromEmbed — FR-010: the /openapi.json handler serves the
// embedded 3.1 contract with 200, independent of any on-disk file (no 503).
func TestServeOpenAPISpec_FromEmbed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/openapi.json", nil)
	rec := httptest.NewRecorder()

	serveOpenAPISpec(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (embedded spec must not 503)", rec.Code)
	}
	var spec struct {
		OpenAPI string                 `json:"openapi"`
		Paths   map[string]any         `json:"paths"`
		Info    map[string]any         `json:"info"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if spec.OpenAPI == "" || spec.OpenAPI[:3] != "3.1" {
		t.Fatalf("expected OpenAPI 3.1.x, got %q", spec.OpenAPI)
	}
	if len(spec.Paths) == 0 {
		t.Fatal("embedded contract has no paths")
	}
}
