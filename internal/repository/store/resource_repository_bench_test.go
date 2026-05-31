package store_test

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/database"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// BenchmarkResourceList_Paired runs Resource.List(100, 0) via both GORM and
// sqlc impls against the same seeded fixture in the same process, gates on
// p95 ratio ≤ 1.10. Spec 049 §FR-006 / SC-002.
//
// Runs on BOTH dialects: SQLite (fast, always-on) + Postgres (gated by
// Docker availability via the existing testcontainer helper). PG bench
// reports the production-relevant ratio; SQLite reports the community-mode
// ratio.
func BenchmarkResourceList_Paired(b *testing.B) {
	b.Run("sqlite", func(b *testing.B) {
		runResourceListPaired(b, benchOpenSQLite(b), "BenchmarkResourceList_Paired/sqlite")
	})
	b.Run("postgres", func(b *testing.B) {
		fx := internaltest.SetupPostgres(b)
		if fx == nil {
			return
		}
		runResourceListPaired(b, fx.Runtime, "BenchmarkResourceList_Paired/postgres")
	})
}

func runResourceListPaired(b *testing.B, rt *database.Runtime, name string) {
	b.Helper()
	gormRepo := store.NewResourceRepository(rt.GormDB())
	sqlcRepo := store.NewResourceRepositorySQLC(rt)
	// Seed via the sqlc resource repo: it only LINKS existing tags /
	// channels through the junction tables, whereas GORM Create with
	// stub associations tries to upsert the channel rows (re-encrypting
	// Config or, for stubs, writing Config=NULL — PG NOT NULL rejects).
	internaltest.SeedPairedBenchFixture(b, internaltest.SeedRepos{
		Tags:      store.NewTagsRepository(rt.GormDB()),
		Channels:  store.NewNotificationChannelRepository(rt.GormDB()),
		Resources: sqlcRepo,
	}, internaltest.FixtureConfig{
		NumResources: 300,
		NumTags:      30,
		NumChannels:  10,
	})
	ctx := context.Background()
	internaltest.RunPairedBench(b, name,
		func() { _, _ = gormRepo.List(ctx, 100, 0) },
		func() { _, _ = sqlcRepo.List(ctx, 100, 0) },
	)
}
