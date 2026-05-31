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
