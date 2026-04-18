package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
)

const (
	maxPerPage     = 100
	defaultPage    = 1
	defaultPerPage = 20
)

// respond writes a single-item envelope response.
func respond(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dtoV1.SingleResponse[interface{}]{
		Data: data,
		Meta: nil,
	})
}

// respondPaginated writes a paginated list response.
func respondPaginated(w http.ResponseWriter, data interface{}, meta dtoV1.MetaResponse) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": data,
		"meta": meta,
	})
}

// respondError writes a structured error response.
func respondError(w http.ResponseWriter, status int, code, message string, fields ...dtoV1.FieldError) {
	detail := dtoV1.ErrorDetail{
		Code:    code,
		Message: message,
	}
	if len(fields) > 0 {
		detail.Fields = fields
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dtoV1.ErrorResponse{Error: detail})
}

// parsePagination parses and validates ?page and ?per_page query params.
// Returns 422 VALIDATION_FAILED for invalid values; clamps per_page > 100 to 100.
func parsePagination(r *http.Request) (dtoV1.PaginationParams, []dtoV1.FieldError) {
	params := dtoV1.PaginationParams{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}
	var errs []dtoV1.FieldError

	if raw := r.URL.Query().Get("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			errs = append(errs, dtoV1.FieldError{Field: "page", Message: "must be a positive integer"})
		} else {
			params.Page = v
		}
	}

	if raw := r.URL.Query().Get("per_page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			errs = append(errs, dtoV1.FieldError{Field: "per_page", Message: "must be a positive integer"})
		} else {
			if v > maxPerPage {
				v = maxPerPage
			}
			params.PerPage = v
		}
	}

	return params, errs
}
