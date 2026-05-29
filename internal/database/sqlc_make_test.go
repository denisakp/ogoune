package database

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSqlcCheckCleanTree shells out to `make sqlc-check` and asserts it
// succeeds on a clean tree. Skipped if `make` is not on PATH (CI runners
// without GNU make, e.g. some hermetic build environments).
func TestSqlcCheckCleanTree(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not on PATH")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	cmd := exec.Command("make", "sqlc-check")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make sqlc-check failed on a clean tree: %v\noutput:\n%s", err, string(out))
	}
}
