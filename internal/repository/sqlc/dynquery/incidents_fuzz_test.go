package dynquery

import (
	"strings"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
)

func FuzzBuildIncidentsQuery(f *testing.F) {
	seeds := []struct {
		status, monitorID string
		fromUnix, toUnix  int64
	}{
		{"open", "'; DROP TABLE incidents; --", 0, 0},
		{"resolved", "' OR '1'='1", 1700000000, 1700000100},
		{"", "%' UNION SELECT * FROM users --", 0, 0},
		{"open", "\x00null", 0, 0},
		{"", strings.Repeat("M", 5000), 0, 0},
		{"", "unicode-';--✓", 0, 0},
		{"resolved", "🔥💀", 0, 0},
		{"", "x'); DELETE FROM incidents WHERE ('1'='1", 1700000000, 1800000000},
		{"open", "\\\\", 0, 0},
		{"", "\";--", 0, 0},
		{"open", strings.Repeat("'", 100), 0, 0},
		{"", "/*comment*/", -1, 1},
		{"resolved", "\t\r\n", 0, 0},
		{"open", "' OR pg_sleep(5) --", 0, 0},
		{"", "x' UNION ALL SELECT NULL,NULL--", 0, 0},
		{"resolved", "🚀", 0, 0},
		{"", "monitor-id-normal", 0, 0},
		{"open", "", 0, 0},
		{"", "漢字", 0, 0},
		{"open", "id;DROP TABLE users", 0, 0},
		{"resolved", "<script>alert(1)</script>", 0, 0},
	}
	for _, s := range seeds {
		f.Add(s.status, s.monitorID, s.fromUnix, s.toUnix)
	}

	f.Fuzz(func(t *testing.T, status, monitorID string, fromUnix, toUnix int64) {
		filter := IncidentFilter{}
		if status != "" {
			filter.Status = &status
		}
		if monitorID != "" {
			filter.MonitorID = &monitorID
		}
		if fromUnix != 0 {
			ft := time.Unix(fromUnix, 0)
			filter.From = &ft
		}
		if toUnix != 0 {
			tt := time.Unix(toUnix, 0)
			filter.To = &tt
		}
		if errs := filter.Validate(); len(errs) > 0 {
			return
		}
		rowsSQL, _, err := BuildIncidentsQuery(filter, 1, 25, sq.Dollar)
		if err != nil {
			return
		}
		countSQL, _, err := BuildIncidentCountQuery(filter, sq.Dollar)
		if err != nil {
			return
		}
		// Real injection signature = quoted literal in SQL.
		for label, payload := range map[string]string{"status": status, "monitor_id": monitorID} {
			if len(payload) < 3 {
				continue
			}
			quoted := "'" + payload + "'"
			if strings.Contains(rowsSQL, quoted) {
				t.Errorf("rows inline-literal leak for %s payload %q: %s", label, payload, rowsSQL)
			}
			if strings.Contains(countSQL, quoted) {
				t.Errorf("count inline-literal leak for %s payload %q: %s", label, payload, countSQL)
			}
		}
	})
}
