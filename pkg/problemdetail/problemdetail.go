package problemdetail

import (
	"encoding/json"
	"net/http"
)

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ProblemDetail implements RFC 7807 structured error responses.
type ProblemDetail struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail"`
	Instance string       `json:"instance,omitempty"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// New creates a ProblemDetail with required fields.
func New(typeURI, title string, status int, detail string) ProblemDetail {
	return ProblemDetail{
		Type:   typeURI,
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// WithInstance sets the instance field (typically the request ID).
func (pd ProblemDetail) WithInstance(instance string) ProblemDetail {
	pd.Instance = instance
	return pd
}

// WithErrors sets the field-level validation errors.
func (pd ProblemDetail) WithErrors(errs []FieldError) ProblemDetail {
	pd.Errors = errs
	return pd
}

// Write sends a ProblemDetail as an application/problem+json response.
func Write(w http.ResponseWriter, pd ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(pd.Status)
	json.NewEncoder(w).Encode(pd)
}
