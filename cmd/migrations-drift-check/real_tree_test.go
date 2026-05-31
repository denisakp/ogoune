package main

import (
	"bytes"
	"testing"
)

// TestRun_RealTree asserts the linter exits 0 against the actual
// internal/database/migrations tree — guards against accidental future drift.
func TestRun_RealTree(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"-root", "../../internal/database/migrations"}, &buf)
	if code != 0 {
		t.Fatalf("real migration tree has drift; linter exit=%d. stderr:\n%s", code, buf.String())
	}
}
