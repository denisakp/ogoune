package dynquery

import (
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func TestBuildMonitorsQuery_NoFilter(t *testing.T) {
	rowsSQL, rowsArgs, countSQL, countArgs, err := BuildMonitorsQuery(MonitorFilter{}, 1, 25, sq.Dollar)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(rowsSQL, "FROM resources r") {
		t.Errorf("expected FROM resources r in rowsSQL, got: %s", rowsSQL)
	}
	if !strings.Contains(rowsSQL, "ORDER BY r.created_at DESC") {
		t.Errorf("missing ORDER BY")
	}
	if !strings.Contains(rowsSQL, "LIMIT 25") {
		t.Errorf("missing LIMIT 25")
	}
	if !strings.Contains(rowsSQL, "OFFSET 0") {
		t.Errorf("missing OFFSET 0")
	}
	if strings.Contains(rowsSQL, "JOIN resource_tags") {
		t.Errorf("unexpected JOIN with no tag filter")
	}
	if len(rowsArgs) != 1 || rowsArgs[0] != true {
		t.Errorf("expected single is_active=true arg, got %v", rowsArgs)
	}
	if !strings.Contains(countSQL, "COUNT(*)") {
		t.Errorf("count SQL missing COUNT(*)")
	}
	if len(countArgs) != 1 {
		t.Errorf("expected 1 count arg, got %v", countArgs)
	}
}

func TestBuildMonitorsQuery_AllFilters(t *testing.T) {
	f := MonitorFilter{
		Tag:      strPtr("production"),
		Type:     strPtr("http"),
		IsActive: boolPtr(false),
		Q:        strPtr("api"),
	}
	rowsSQL, rowsArgs, _, _, err := BuildMonitorsQuery(f, 2, 10, sq.Dollar)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(rowsSQL, "JOIN resource_tags rt") {
		t.Errorf("expected tag JOIN, got: %s", rowsSQL)
	}
	if !strings.Contains(rowsSQL, "JOIN tags t") {
		t.Errorf("expected tags JOIN")
	}
	if !strings.Contains(rowsSQL, "LIMIT 10") || !strings.Contains(rowsSQL, "OFFSET 10") {
		t.Errorf("bad pagination: %s", rowsSQL)
	}
	// Values must appear as ARGS, never inline.
	for _, v := range []string{"production", "http", "api"} {
		if strings.Contains(rowsSQL, "'"+v+"'") {
			t.Errorf("value %q inlined in SQL: %s", v, rowsSQL)
		}
	}
	if !containsAll(rowsArgs, false, "http", "production") {
		t.Errorf("args missing values: %v", rowsArgs)
	}
}

func TestBuildMonitorsQuery_IsActiveOverride(t *testing.T) {
	// When IsActive is explicitly false, default TRUE must not appear.
	f := MonitorFilter{IsActive: boolPtr(false)}
	_, rowsArgs, _, _, err := BuildMonitorsQuery(f, 1, 25, sq.Dollar)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for _, a := range rowsArgs {
		if a == true {
			t.Errorf("default is_active=TRUE leaked when explicit false passed: %v", rowsArgs)
		}
	}
}

func TestBuildMonitorsQuery_NeverInlinesValues(t *testing.T) {
	tag := "'; DROP TABLE resources; --"
	qq := "%admin%"
	rowsSQL, rowsArgs, countSQL, _, err := BuildMonitorsQuery(
		MonitorFilter{Tag: &tag, Q: &qq},
		1, 25, sq.Dollar,
	)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if strings.Contains(rowsSQL, "DROP TABLE") {
		t.Errorf("SQL injection leaked: %s", rowsSQL)
	}
	if strings.Contains(rowsSQL, "admin") {
		t.Errorf("Q value leaked into rowsSQL: %s", rowsSQL)
	}
	if strings.Contains(countSQL, "DROP TABLE") {
		t.Errorf("SQL injection leaked into countSQL: %s", countSQL)
	}
	if !containsAll(rowsArgs, tag) {
		t.Errorf("tag value missing from args: %v", rowsArgs)
	}
}

func TestBuildMonitorsQuery_PlaceholderDialects(t *testing.T) {
	f := MonitorFilter{Type: strPtr("http")}
	pg, _, _, _, _ := BuildMonitorsQuery(f, 1, 25, sq.Dollar)
	lite, _, _, _, _ := BuildMonitorsQuery(f, 1, 25, sq.Question)
	if !strings.Contains(pg, "$1") {
		t.Errorf("expected $N placeholders for Postgres, got: %s", pg)
	}
	if strings.Contains(lite, "$1") || !strings.Contains(lite, "?") {
		t.Errorf("expected ? placeholders for SQLite, got: %s", lite)
	}
}

func TestBuildMonitorsQuery_LongQ(t *testing.T) {
	long := strings.Repeat("x", 200)
	f := MonitorFilter{Q: &long}
	_, args, _, _, err := BuildMonitorsQuery(f, 1, 25, sq.Dollar)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected := "%" + long + "%"
	if !containsAll(args, expected) {
		t.Errorf("escaped long Q missing from args")
	}
}

func TestMonitorFilter_Validate(t *testing.T) {
	cases := []struct {
		name     string
		f        MonitorFilter
		wantErrs int
	}{
		{"empty ok", MonitorFilter{}, 0},
		{"valid type", MonitorFilter{Type: strPtr("http")}, 0},
		{"bad type", MonitorFilter{Type: strPtr("bogus")}, 1},
		{"q too long", MonitorFilter{Q: strPtr(strings.Repeat("x", 201))}, 1},
		{"tag too long", MonitorFilter{Tag: strPtr(strings.Repeat("t", 101))}, 1},
		{"is_active does not validate value", MonitorFilter{IsActive: boolPtr(true)}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := c.f.Validate()
			if len(errs) != c.wantErrs {
				t.Errorf("want %d errs, got %d: %v", c.wantErrs, len(errs), errs)
			}
		})
	}
}

func containsAll(args []any, want ...any) bool {
	for _, w := range want {
		found := false
		for _, a := range args {
			if a == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
