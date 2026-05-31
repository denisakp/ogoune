package database

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestSqlcCheckDetectsDrift mutates a pilot query, runs `make sqlc-check`,
// and asserts it exits non-zero with the expected drift message. The
// original file is restored on cleanup so the working tree stays clean.
func TestSqlcCheckDetectsDrift(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not on PATH")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	pilot := filepath.Join(repoRoot, "internal", "repository", "sqlc", "queries", "sqlite", "ping.sql")

	original, err := os.ReadFile(pilot)
	if err != nil {
		t.Fatalf("read pilot query: %v", err)
	}
	t.Cleanup(func() {
		if writeErr := os.WriteFile(pilot, original, 0o644); writeErr != nil {
			t.Errorf("restore pilot query: %v", writeErr)
		}
		// Regenerate to restore committed generated files to their tracked state.
		_ = exec.Command("make", "-C", repoRoot, "sqlc-generate").Run()
	})

	// Rename the query so generated Go function name changes, producing a diff.
	mutated := strings.ReplaceAll(string(original), "name: Ping", "name: Pong")
	if err := os.WriteFile(pilot, []byte(mutated), 0o644); err != nil {
		t.Fatalf("mutate pilot query: %v", err)
	}

	cmd := exec.Command("make", "sqlc-check")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected make sqlc-check to fail on drift, got success\noutput:\n%s", string(out))
	}
	if !strings.Contains(string(out), "sqlc drift") {
		t.Errorf("expected 'sqlc drift' in output, got:\n%s", string(out))
	}
}
