package dynquery

import (
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func FuzzBuildMonitorsQuery(f *testing.F) {
	seeds := []struct {
		tag, typ, q string
		active      bool
	}{
		{"'; DROP TABLE resources; --", "http", "", true},
		{"' OR '1'='1", "tcp", "' OR 1=1 --", false},
		{"%' UNION SELECT * FROM users --", "dns", "anything", true},
		{"\x00null-byte", "http", "\x00", true},
		{strings.Repeat("A", 5000), "http", strings.Repeat("B", 5000), false},
		{"unicode-';--✓", "icmp", "日本語'; DROP", true},
		{"", "", "%_\\admin", true},
		{"normal-tag", "keyword", "search-me", true},
		{"\n--comment\nDROP", "protocol", "/*comment*/", false},
		{"`backtick`", "heartbeat", "name=\"x\"", true},
		{"'); DELETE FROM resources WHERE ('1'='1", "http", "", true},
		{"\\", "tcp", "\\\\", false},
		{strings.Repeat("'", 100), "dns", strings.Repeat(";", 100), true},
		{"%", "icmp", "_", true},
		{"\";--", "http", "\"; DROP --", false},
		{"\t\r\n", "tcp", " ", true},
		{"\\x27\\x22", "dns", "\\u0027", true},
		{"' OR sleep(10) --", "http", "' AND BENCHMARK(1000000,MD5('x')) --", true},
		{"x' UNION ALL SELECT NULL,NULL--", "tcp", "", false},
		{"' OR 1=CAST((SELECT pg_sleep(5)) AS TEXT) --", "http", "", true},
		{"漢字漢字", "http", "🔥💀", true},
	}
	for _, s := range seeds {
		f.Add(s.tag, s.typ, s.q, s.active)
	}

	f.Fuzz(func(t *testing.T, tag, typ, q string, active bool) {
		filter := MonitorFilter{Tag: &tag, Type: &typ, Q: &q, IsActive: &active}
		// Validate may reject — that's fine, no SQL is built then.
		if errs := filter.Validate(); len(errs) > 0 {
			return
		}
		rowsSQL, _, countSQL, _, err := BuildMonitorsQuery(filter, 1, 25, sq.Dollar)
		if err != nil {
			return
		}
		// Payload must never appear inline as a quoted literal in SQL.
		// We check for the raw payload — any presence is suspicious unless
		// matched only by the column / keyword text.
		// Real injection signature = payload wrapped as a quoted SQL literal.
		// Single chars that overlap with SQL syntax (`_`, ` `, `\`) appear
		// legitimately in operators/keywords and are not injection vectors.
		for label, payload := range map[string]string{"tag": tag, "type": typ, "q": q} {
			if len(payload) < 3 {
				continue
			}
			quoted := "'" + payload + "'"
			if strings.Contains(rowsSQL, quoted) {
				t.Errorf("rows SQL inline-literal leak for %s payload %q: %s", label, payload, rowsSQL)
			}
			if strings.Contains(countSQL, quoted) {
				t.Errorf("count SQL inline-literal leak for %s payload %q: %s", label, payload, countSQL)
			}
		}
	})
}
