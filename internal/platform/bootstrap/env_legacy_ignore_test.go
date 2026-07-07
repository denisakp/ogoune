package bootstrap

import (
	"bytes"
	"context"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	dbruntime "github.com/denisakp/ogoune/internal/database"
)

// TestLegacyFlagsSilentlyIgnored is the FR-004 / contracts/env-contract.md
// regression guard. Post-spec-052, every legacy SQLC_* env var must be
// silently ignored by the binary — present or absent, valid or garbage,
// the boot path's observable behaviour is identical and no log line ever
// mentions an SQLC_* flag.
//
// We exercise database init (the file that used to read these flags) and
// scrape both stdout/stderr via slog redirection. Anything with the
// regex `SQLC_[A-Z]` in the output = bug.
func TestLegacyFlagsSilentlyIgnored(t *testing.T) {
	allLegacyFlags := []string{
		"SQLC_TAGS",
		"SQLC_API_KEY",
		"SQLC_USER",
		"SQLC_NOTIFICATION_CHANNEL",
		"SQLC_EXPIRY_NOTIFICATION_LOG",
		"SQLC_STATUSPAGE_SETTINGS",
		"SQLC_INCIDENT_DIAGNOSTICS",
		"SQLC_RESOURCE_CREDENTIAL",
		"SQLC_COMPONENT",
		"SQLC_MAINTENANCE",
		"SQLC_NOTIFICATION",
		"SQLC_MONITORING_ACTIVITY",
		"SQLC_INCIDENT_EVENT_STEP",
		"SQLC_RESOURCE",
		"SQLC_INCIDENT",
		"SQLC_NEWTHING", // future-namespace canary
	}

	type envCase struct {
		name string
		env  map[string]string
	}

	cases := []envCase{
		{"baseline_empty", map[string]string{}},
		{"all_true", flagsSet(allLegacyFlags, "true")},
		{"all_false", flagsSet(allLegacyFlags, "false")},
		{"all_garbage", flagsSet(allLegacyFlags, "banana")},
		{"mixed", map[string]string{
			"SQLC_TAGS":     "true",
			"SQLC_RESOURCE": "false",
			"SQLC_INCIDENT": "garbage",
		}},
	}

	sqlcLineRE := regexp.MustCompile(`SQLC_[A-Z]`)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for k, v := range c.env {
				t.Setenv(k, v)
			}

			// Capture slog output during the boot path.
			var buf bytes.Buffer
			prev := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
			t.Cleanup(func() { slog.SetDefault(prev) })

			// Open a fresh SQLite runtime (this is the path that used to read
			// SQLC_* flags; we exercise it end-to-end). Direct dbruntime call
			// avoids importing config / icmp / crypto deps.
			dbPath := filepath.Join(t.TempDir(), "legacy-ignore.db")
			rt, err := dbruntime.Open(context.Background(), dbruntime.Config{
				Driver:     dbruntime.DriverSQLite,
				SQLitePath: dbPath,
				LogLevel:   "silent",
			})
			if err != nil {
				t.Fatalf("boot failed: %v", err)
			}
			if rt == nil {
				t.Fatalf("nil runtime")
			}
			t.Cleanup(func() {
				if rt.SQLiteDB() != nil {
					_ = rt.SQLiteDB().Close()
				}
			})

			out := buf.String()
			if loc := sqlcLineRE.FindStringIndex(out); loc != nil {
				// Pick the offending line for the diagnostic.
				line := out[:loc[1]]
				if i := strings.LastIndex(line, "\n"); i >= 0 {
					line = strings.TrimSpace(out[i+1 : loc[1]+200])
				}
				t.Fatalf("log mentioned SQLC_* flag — should be silent. Line: %q\nFull output:\n%s", line, out)
			}
		})
	}
}

func flagsSet(keys []string, val string) map[string]string {
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = val
	}
	return out
}
