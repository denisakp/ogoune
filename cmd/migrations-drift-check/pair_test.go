package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_OK(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"-root", "testdata/ok"}, &buf)
	if code != 0 {
		t.Fatalf("expected exit 0 on ok fixture, got %d. stderr:\n%s", code, buf.String())
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty stderr on ok fixture, got:\n%s", buf.String())
	}
}

func TestRun_MissingPair(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"-root", "testdata/missing_pair"}, &buf)
	if code != 1 {
		t.Fatalf("expected exit 1 on missing_pair fixture, got %d. stderr:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "missing pair for prefix 0001") {
		t.Errorf("expected 'missing pair for prefix 0001' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "sqlite=(missing)") {
		t.Errorf("expected 'sqlite=(missing)' in output, got:\n%s", out)
	}
}

// Two migrations sharing a numeric prefix are indistinguishable to the
// migrator: applied state is keyed on that prefix alone, so the second is
// skipped forever on any database that already recorded the number. The
// migrator refuses to start on it; this catches it in CI first.
func TestRun_DuplicatePrefix(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"-root", "testdata/duplicate_prefix"}, &buf)
	if code == 0 {
		t.Fatalf("expected a non-zero exit on a duplicate prefix, got 0")
	}
	out := buf.String()
	if !strings.Contains(out, "share prefix 0001") {
		t.Errorf("expected 'share prefix 0001' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "every migration needs its own number") {
		t.Errorf("the message must say what to do about it, got:\n%s", out)
	}
}

// The checker used to match only `NNNN_name.sql`, whose name part cannot contain
// a dot — so every `NNNN_name.up.sql` was skipped, which was 40 of the 55 files
// in the real tree. It reported success while inspecting a quarter of it.
//
// This fixture is a paired migration with a genuine nullability drift between
// dialects. Passing it means the guard is blind again.
func TestRun_SeesPairedUpFiles(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"-root", "testdata/paired_files"}, &buf)
	if code != 1 {
		t.Fatalf("expected exit 1 on a drifting paired migration, got %d. stderr:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "table=foo, column=name") {
		t.Errorf("expected the drifting column to be named, got:\n%s", out)
	}
}
