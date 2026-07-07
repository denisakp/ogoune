package v1

import (
	"net/http"
	"strconv"
	"strings"

	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
)

// parseMonitorFilter extracts dynamic filter query params for GET /monitors.
// Empty params map to nil pointers (= no filter on that field).
func parseMonitorFilter(r *http.Request) (dynquery.MonitorFilter, []dtoV1.FieldError) {
	q := r.URL.Query()
	var f dynquery.MonitorFilter
	var errs []dtoV1.FieldError

	if v := strings.TrimSpace(q.Get("tag")); v != "" {
		f.Tag = &v
	}
	if v := strings.TrimSpace(q.Get("type")); v != "" {
		f.Type = &v
	}
	if v := strings.TrimSpace(q.Get("is_active")); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, dtoV1.FieldError{Field: "is_active", Message: "must be true/false/1/0"})
		} else {
			f.IsActive = &b
		}
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		f.Q = &v
	}
	if ve := f.Validate(); len(ve) > 0 {
		errs = append(errs, ve...)
	}
	return f, errs
}
