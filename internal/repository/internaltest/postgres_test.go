package internaltest_test

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestSetupPostgres_TemplateCloneIsolates verifies the Postgres helper
// boots a container (or uses POSTGRES_TEST_DSN), applies migrations to a
// template, and gives each test an isolated database. Skips if neither
// Docker nor POSTGRES_TEST_DSN is available.
func TestSetupPostgres_TemplateCloneIsolates(t *testing.T) {
	fxA := internaltest.SetupPostgres(t)
	if fxA == nil {
		return // SetupPostgres already called t.Skip
	}
	if fxA.Dialect != "postgres" {
		t.Fatalf("expected dialect postgres, got %q", fxA.Dialect)
	}
	if fxA.DSN == "" {
		t.Error("expected non-empty DSN on Postgres fixture")
	}

	ctx := context.Background()
	tagsA := store.NewTagsRepositorySQLC(fxA.Runtime)
	tag := &domain.Tags{Name: "pg-isolation-a"}
	tag.EnsureID()
	if err := tagsA.Create(ctx, tag); err != nil {
		t.Fatalf("create in fixture A: %v", err)
	}

	// A second sub-test gets a fresh DB cloned from the template.
	t.Run("fresh_db", func(t *testing.T) {
		fxB := internaltest.SetupPostgres(t)
		if fxB == nil {
			return
		}
		// Raw COUNT(*) via pgx pool — the template clone must produce 0 rows.
		var count int
		if err := fxB.Runtime.PgxPool().QueryRow(ctx, "SELECT COUNT(*) FROM tags").Scan(&count); err != nil {
			t.Fatalf("count B: %v", err)
		}
		if count != 0 {
			t.Errorf("expected fresh DB clone to have 0 rows, got %d", count)
		}
	})
}
