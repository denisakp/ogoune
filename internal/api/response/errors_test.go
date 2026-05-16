package response

import (
	"net/http"
	"testing"

	"github.com/denisakp/ogoune/internal/service"
)

func TestMapServiceError(t *testing.T) {
	tests := []struct {
		err    error
		status int
		typ    string
		title  string
	}{
		{service.ErrValidationFailed, http.StatusUnprocessableEntity, "/problems/validation-failed", "Validation Failed"},
		{service.ErrResourceNotFound, http.StatusNotFound, "/problems/not-found", "Resource Not Found"},
		{service.ErrInvalidCredentials, http.StatusUnauthorized, "/problems/invalid-credentials", "Invalid Credentials"},
		{service.ErrUnauthorized, http.StatusUnauthorized, "/problems/unauthorized", "Unauthorized"},
		{service.ErrInvalidToken, http.StatusUnauthorized, "/problems/invalid-token", "Invalid Token"},
		{service.ErrInvalidPassword, http.StatusUnprocessableEntity, "/problems/validation-failed", "Validation Failed"},
		{service.ErrAPIKeyNotFound, http.StatusNotFound, "/problems/not-found", "API Key Not Found"},
		{service.ErrAPIKeyLimitReached, http.StatusUnprocessableEntity, "/problems/limit-reached", "Limit Reached"},
		{service.ErrAPIKeyExpired, http.StatusUnauthorized, "/problems/key-expired", "API Key Expired"},
		{service.ErrAPIKeyRevoked, http.StatusUnauthorized, "/problems/key-revoked", "API Key Revoked"},
		{service.ErrSchedulerSync, http.StatusInternalServerError, "/problems/internal-error", "Internal Error"},
		{service.ErrICMPUnavailable, http.StatusServiceUnavailable, "/problems/service-unavailable", "Service Unavailable"},
		{service.ErrMaintenanceNotFound, http.StatusNotFound, "/problems/not-found", "Maintenance Not Found"},
	}

	for _, tc := range tests {
		t.Run(tc.err.Error(), func(t *testing.T) {
			pd := MapServiceError(tc.err)

			if pd.Status != tc.status {
				t.Errorf("status: got %d, want %d", pd.Status, tc.status)
			}
			if pd.Type != tc.typ {
				t.Errorf("type: got %q, want %q", pd.Type, tc.typ)
			}
			if pd.Title != tc.title {
				t.Errorf("title: got %q, want %q", pd.Title, tc.title)
			}
			if pd.Detail == "" {
				t.Error("detail should not be empty")
			}
		})
	}
}

func TestMapServiceError_UnknownError(t *testing.T) {
	pd := MapServiceError(http.ErrServerClosed) // random unrecognized error

	if pd.Status != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", pd.Status)
	}
	if pd.Type != "/problems/internal-error" {
		t.Errorf("type: got %q", pd.Type)
	}
}
