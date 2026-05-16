package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/denisakp/ogoune/pkg/problemdetail"
)

// JSON writes a JSON response with the given status code and payload.
// This is a centralized helper to ensure consistent JSON response formatting.
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			// If encoding fails after headers are written, log it but can't send error response
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// Error writes an RFC 7807 ProblemDetail error response.
func Error(w http.ResponseWriter, status int, message string) {
	pd := problemdetail.New(
		fmt.Sprintf("/problems/%s", problemTypeFromStatus(status)),
		http.StatusText(status),
		status,
		message,
	)
	problemdetail.Write(w, pd)
}

// ErrorWithRequest writes an RFC 7807 ProblemDetail error response with instance (request ID).
func ErrorWithRequest(w http.ResponseWriter, r *http.Request, status int, message string) {
	pd := problemdetail.New(
		fmt.Sprintf("/problems/%s", problemTypeFromStatus(status)),
		http.StatusText(status),
		status,
		message,
	)
	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		pd = pd.WithInstance(reqID)
	}
	problemdetail.Write(w, pd)
}

// problemTypeFromStatus returns a URL-safe problem type slug for an HTTP status code.
func problemTypeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad-request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not-found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusUnprocessableEntity:
		return "validation-failed"
	case http.StatusTooManyRequests:
		return "rate-limit-exceeded"
	case http.StatusServiceUnavailable:
		return "service-unavailable"
	default:
		return "internal-error"
	}
}

// Success writes a JSON success response with a message.
// Success responses follow the format: {"message": "success message"}
func Success(w http.ResponseWriter, message string) {
	JSON(w, http.StatusOK, map[string]string{"message": message})
}

// Created writes a JSON response for resource creation (201 Created).
func Created(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusCreated, payload)
}

// NoContent writes a 204 No Content response (empty body).
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
