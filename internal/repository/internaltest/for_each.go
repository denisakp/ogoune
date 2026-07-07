package internaltest

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ForEachDialect runs fn once per supported dialect under named sub-tests.
// The "sqlite" iteration ALWAYS runs. The "postgres" iteration runs when
// either POSTGRES_TEST_DSN is set or a Docker daemon is reachable; otherwise
// SetupPostgres calls t.Skip and the iteration is skipped — but the SQLite
// iteration is unaffected.
func ForEachDialect(t *testing.T, fn func(t *testing.T, fx *DialectFixture)) {
	t.Helper()

	t.Run("sqlite", func(t *testing.T) {
		fn(t, SetupSQLite(t))
	})
	t.Run("postgres", func(t *testing.T) {
		fx := SetupPostgres(t)
		if fx == nil {
			return // t.Skip already called
		}
		fn(t, fx)
	})
}

// DialectsAvailable reports which dialects will actually execute in the
// current environment. Useful for CI introspection and test-log clarity.
func DialectsAvailable() []string {
	out := []string{"sqlite"}
	if strings.TrimSpace(os.Getenv("POSTGRES_TEST_DSN")) != "" || dockerReachable() {
		out = append(out, "postgres")
	}
	return out
}

func dockerReachable() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	cmd := exec.Command("docker", "info")
	cmd.Stdout, cmd.Stderr = nil, nil
	return cmd.Run() == nil
}
