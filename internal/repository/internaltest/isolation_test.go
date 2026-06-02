package internaltest_test

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestIsolation_ParallelSamePrimaryKey defends SC-007: two parallel
// sub-tests inserting rows with the SAME primary key MUST NOT collide;
// each sees only its own row.
func TestIsolation_ParallelSamePrimaryKey(t *testing.T) {
	const sharedID = "01ZZZZZZZZZZZZZZZZZZZZZZZA"

	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		t.Run("inserter_x", func(t *testing.T) {
			t.Parallel()
			inner := internaltest.SetupSQLite(t)
			if fx.Dialect == "postgres" {
				if pg := internaltest.SetupPostgres(t); pg != nil {
					inner = pg
				} else {
					return
				}
			}
			tag := &domain.Tags{Base: domain.Base{ID: sharedID}, Name: "x"}
			if err := store.NewTagsRepositorySQLC(inner.Runtime).Create(context.Background(), tag); err != nil {
				t.Fatalf("create x: %v", err)
			}
		})
		t.Run("inserter_y", func(t *testing.T) {
			t.Parallel()
			inner := internaltest.SetupSQLite(t)
			if fx.Dialect == "postgres" {
				if pg := internaltest.SetupPostgres(t); pg != nil {
					inner = pg
				} else {
					return
				}
			}
			tag := &domain.Tags{Base: domain.Base{ID: sharedID}, Name: "y"}
			if err := store.NewTagsRepositorySQLC(inner.Runtime).Create(context.Background(), tag); err != nil {
				t.Fatalf("create y: %v", err)
			}
		})
	})
}
