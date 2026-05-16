package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/pkg/logger"
	"github.com/denisakp/ogoune/pkg/problemdetail"
)

const (
	maxPerPage     = 100
	defaultPage    = 1
	defaultPerPage = 20
)

// respond writes a single-item envelope response.
func respond(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dtoV1.SingleResponse[any]{
		Data: data,
		Meta: nil,
	})
}

// respondPaginated writes a paginated list response.
func respondPaginated(w http.ResponseWriter, data any, meta dtoV1.MetaResponse) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": data,
		"meta": meta,
	})
}

// respondError writes an RFC 7807 ProblemDetail error response.
func respondError(w http.ResponseWriter, r *http.Request, status int, code, message string, fields ...dtoV1.FieldError) {
	pd := problemdetail.New(code, http.StatusText(status), status, message)

	if reqID := logger.RequestID(r.Context()); reqID != "" {
		pd = pd.WithInstance(reqID)
	}

	if len(fields) > 0 {
		pdErrors := make([]problemdetail.FieldError, len(fields))
		for i, f := range fields {
			pdErrors[i] = problemdetail.FieldError{
				Field:   f.Field,
				Message: f.Message,
			}
		}
		pd = pd.WithErrors(pdErrors)
	}

	problemdetail.Write(w, pd)
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
