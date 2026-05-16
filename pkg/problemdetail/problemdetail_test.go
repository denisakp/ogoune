package problemdetail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_SetsRequiredFields(t *testing.T) {
	pd := New("/problems/not-found", "Resource Not Found", 404, "monitor xyz not found")

	if pd.Type != "/problems/not-found" {
		t.Errorf("Type = %q", pd.Type)
	}
	if pd.Title != "Resource Not Found" {
		t.Errorf("Title = %q", pd.Title)
	}
	if pd.Status != 404 {
		t.Errorf("Status = %d", pd.Status)
	}
	if pd.Detail != "monitor xyz not found" {
		t.Errorf("Detail = %q", pd.Detail)
	}
}

func TestWithInstance(t *testing.T) {
	pd := New("/problems/not-found", "Not Found", 404, "detail").
		WithInstance("01HYABCDEF123")

	if pd.Instance != "01HYABCDEF123" {
		t.Errorf("Instance = %q", pd.Instance)
	}
}

func TestWithErrors(t *testing.T) {
	errs := []FieldError{
		{Field: "name", Message: "is required"},
		{Field: "url", Message: "must be a valid URL", Code: "invalid_format"},
	}
	pd := New("/problems/validation-failed", "Validation Failed", 422, "request validation failed").
		WithErrors(errs)

	if len(pd.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(pd.Errors))
	}
	if pd.Errors[0].Field != "name" {
		t.Errorf("first error field = %q", pd.Errors[0].Field)
	}
	if pd.Errors[1].Code != "invalid_format" {
		t.Errorf("second error code = %q", pd.Errors[1].Code)
	}
}

func TestWrite_ContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	pd := New("/problems/not-found", "Not Found", 404, "not found")
	Write(rec, pd)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/problem+json" {
		t.Errorf("Content-Type = %q", ct)
	}

	if rec.Code != 404 {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestWrite_JSONOutput_MatchesSchema(t *testing.T) {
	rec := httptest.NewRecorder()
	pd := New("/problems/validation-failed", "Validation Failed", 422, "invalid input").
		WithInstance("req-123").
		WithErrors([]FieldError{
			{Field: "email", Message: "is required"},
		})

	Write(rec, pd)

	var got map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Required fields
	for _, field := range []string{"type", "title", "status", "detail"} {
		if _, ok := got[field]; !ok {
			t.Errorf("missing required field %q", field)
		}
	}

	if got["status"].(float64) != 422 {
		t.Errorf("status = %v", got["status"])
	}

	if got["instance"] != "req-123" {
		t.Errorf("instance = %v", got["instance"])
	}

	errs, ok := got["errors"].([]interface{})
	if !ok || len(errs) != 1 {
		t.Fatalf("expected 1 error, got %v", got["errors"])
	}
}

func TestWrite_OmitsEmptyOptionalFields(t *testing.T) {
	rec := httptest.NewRecorder()
	pd := New("/problems/not-found", "Not Found", 404, "not found")
	Write(rec, pd)

	var got map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &got)

	if _, ok := got["instance"]; ok {
		t.Error("instance should be omitted when empty")
	}
	if _, ok := got["errors"]; ok {
		t.Error("errors should be omitted when nil")
	}
}

func TestWrite_StatusCodes(t *testing.T) {
	tests := []struct {
		status int
	}{
		{http.StatusBadRequest},
		{http.StatusUnauthorized},
		{http.StatusNotFound},
		{http.StatusUnprocessableEntity},
		{http.StatusInternalServerError},
		{http.StatusServiceUnavailable},
	}

	for _, tc := range tests {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			pd := New("/problems/test", "Test", tc.status, "test")
			Write(rec, pd)

			if rec.Code != tc.status {
				t.Errorf("expected %d, got %d", tc.status, rec.Code)
			}
		})
	}
}
