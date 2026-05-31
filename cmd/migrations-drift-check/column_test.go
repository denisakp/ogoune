package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_NullDrift(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"-root", "testdata/null_drift"}, &buf)
	if code != 1 {
		t.Fatalf("expected exit 1 on null_drift fixture, got %d. stderr:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "nullability drift") {
		t.Errorf("expected 'nullability drift' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "column=note") {
		t.Errorf("expected 'column=note' in output, got:\n%s", out)
	}
}
