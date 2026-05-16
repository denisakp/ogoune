package response

import (
	"errors"
	"net/http"

	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/pkg/problemdetail"
)

const problemNotFound = "/problems/not-found"

type errorMapping struct {
	Status int
	Type   string
	Title  string
}

var mappings = map[error]errorMapping{
	service.ErrValidationFailed:    {http.StatusUnprocessableEntity, "/problems/validation-failed", "Validation Failed"},
	service.ErrResourceNotFound:    {http.StatusNotFound, problemNotFound, "Resource Not Found"},
	service.ErrInvalidCredentials:  {http.StatusUnauthorized, "/problems/invalid-credentials", "Invalid Credentials"},
	service.ErrUnauthorized:        {http.StatusUnauthorized, "/problems/unauthorized", "Unauthorized"},
	service.ErrInvalidToken:        {http.StatusUnauthorized, "/problems/invalid-token", "Invalid Token"},
	service.ErrInvalidPassword:     {http.StatusUnprocessableEntity, "/problems/validation-failed", "Validation Failed"},
	service.ErrAPIKeyNotFound:      {http.StatusNotFound, problemNotFound, "API Key Not Found"},
	service.ErrAPIKeyLimitReached:  {http.StatusUnprocessableEntity, "/problems/limit-reached", "Limit Reached"},
	service.ErrAPIKeyExpired:       {http.StatusUnauthorized, "/problems/key-expired", "API Key Expired"},
	service.ErrAPIKeyRevoked:       {http.StatusUnauthorized, "/problems/key-revoked", "API Key Revoked"},
	service.ErrAPIKeyInvalid:       {http.StatusUnauthorized, "/problems/key-revoked", "API Key Revoked"},
	service.ErrSchedulerSync:       {http.StatusInternalServerError, "/problems/internal-error", "Internal Error"},
	service.ErrICMPUnavailable:     {http.StatusServiceUnavailable, "/problems/service-unavailable", "Service Unavailable"},
	service.ErrMaintenanceNotFound: {http.StatusNotFound, problemNotFound, "Maintenance Not Found"},
}

// MapServiceError converts a service-layer error to a ProblemDetail.
func MapServiceError(err error) problemdetail.ProblemDetail {
	for sentinel, m := range mappings {
		if errors.Is(err, sentinel) {
			return problemdetail.New(m.Type, m.Title, m.Status, err.Error())
		}
	}
	return problemdetail.New("/problems/internal-error", "Internal Error", http.StatusInternalServerError, "an unexpected error occurred")
}
