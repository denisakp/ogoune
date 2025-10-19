package response

import (
	"encoding/json"
	"net/http"
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

// Error writes a JSON error response with the given status code and message.
// Error responses follow a consistent format: {"error": "message"}
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
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
