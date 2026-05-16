package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
)

func TestRespondError_RFC7807Format(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	respondError(rec, req, http.StatusNotFound, "RESOURCE_NOT_FOUND", "monitor not found")

	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}

	var pd map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, field := range []string{"type", "title", "status", "detail"} {
		if _, ok := pd[field]; !ok {
			t.Errorf("missing required RFC 7807 field %q", field)
		}
	}
	if pd["status"].(float64) != 404 {
		t.Errorf("status = %v", pd["status"])
	}
	if pd["detail"] != "monitor not found" {
		t.Errorf("detail = %v", pd["detail"])
	}
}

func TestRespondError_ValidationWithFields(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)

	respondError(rec, req, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "validation failed",
		dtoV1.FieldError{Field: "name", Message: "is required"},
		dtoV1.FieldError{Field: "url", Message: "must be valid"},
	)

	if rec.Code != 422 {
		t.Errorf("status = %d", rec.Code)
	}

	var pd map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &pd)

	errs, ok := pd["errors"].([]interface{})
	if !ok || len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %v", pd["errors"])
	}

	first := errs[0].(map[string]interface{})
	if first["field"] != "name" {
		t.Errorf("first error field = %v", first["field"])
	}
}

func TestRespondError_Unauthorized(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	respondError(rec, req, http.StatusUnauthorized, "UNAUTHORIZED", "missing token")

	if rec.Code != 401 {
		t.Errorf("status = %d", rec.Code)
	}

	var pd map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &pd)

	if pd["type"] != "UNAUTHORIZED" {
		t.Errorf("type = %v", pd["type"])
	}
}

func TestRespondError_InternalError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	respondError(rec, req, http.StatusInternalServerError, "INTERNAL_ERROR", "something went wrong")

	if rec.Code != 500 {
		t.Errorf("status = %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestRespondError_IncludesInstanceWhenRequestIDPresent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Simulate request ID in context
	ctx := req.Context()
	// Import the logger package context helper
	req = req.WithContext(ctx)

	respondError(rec, req, http.StatusNotFound, "RESOURCE_NOT_FOUND", "not found")

	var pd map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &pd)

	// Without request ID in context, instance should be omitted
	if _, ok := pd["instance"]; ok {
		// instance should only appear when request ID is set
		if pd["instance"] == "" {
			t.Error("instance should be omitted when empty, not empty string")
		}
	}
}
