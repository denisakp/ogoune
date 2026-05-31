package internaltest_test

import (
	"sync/atomic"
	"testing"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
)

func TestForEachDialect_RunsBothAndIsolates(t *testing.T) {
	var sqliteRan, postgresRan atomic.Bool

	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		switch fx.Dialect {
		case "sqlite":
			sqliteRan.Store(true)
		case "postgres":
			postgresRan.Store(true)
		default:
			t.Fatalf("unexpected dialect: %q", fx.Dialect)
		}
		if fx.Runtime == nil {
			t.Fatalf("fixture for %s has nil Runtime", fx.Dialect)
		}
	})

	if !sqliteRan.Load() {
		t.Error("SQLite iteration MUST run regardless of Postgres availability")
	}
	// postgres iteration is skip-aware; we only check it did NOT explode.
}

func TestDialectsAvailable_AlwaysIncludesSQLite(t *testing.T) {
	dialects := internaltest.DialectsAvailable()
	found := false
	for _, d := range dialects {
		if d == "sqlite" {
			found = true
		}
	}
	if !found {
		t.Errorf("DialectsAvailable() must always include sqlite; got %v", dialects)
	}
}
