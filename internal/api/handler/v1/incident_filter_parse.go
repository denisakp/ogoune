package v1

import (
	"net/http"
	"strings"
	"time"

	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/repository/sqlc/dynquery"
)

// parseIncidentFilter extracts dynamic filter query params for GET /incidents.
// Empty params map to nil pointers. `from` / `to` accept RFC 3339.
func parseIncidentFilter(r *http.Request) (dynquery.IncidentFilter, []dtoV1.FieldError) {
	q := r.URL.Query()
	var f dynquery.IncidentFilter
	var errs []dtoV1.FieldError

	if v := strings.TrimSpace(q.Get("status")); v != "" {
		f.Status = &v
	}
	if v := strings.TrimSpace(q.Get("monitor_id")); v != "" {
		f.MonitorID = &v
	}
	if v := strings.TrimSpace(q.Get("from")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			errs = append(errs, dtoV1.FieldError{Field: "from", Message: "must be RFC 3339"})
		} else {
			f.From = &t
		}
	}
	if v := strings.TrimSpace(q.Get("to")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			errs = append(errs, dtoV1.FieldError{Field: "to", Message: "must be RFC 3339"})
		} else {
			f.To = &t
		}
	}
	if ve := f.Validate(); len(ve) > 0 {
		errs = append(errs, ve...)
	}
	return f, errs
}
