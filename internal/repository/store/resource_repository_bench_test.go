package store_test

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// BenchmarkResourceList_Paired runs Resource.List(100, 0) via both GORM and
// sqlc impls against the same seeded fixture in the same process, gates on
// p95 ratio ≤ 1.10. Spec 049 §FR-006 / SC-002.
//
// SQLite only: paired benches don't need cross-dialect coverage (the ratio
// is the signal, dialect-invariant). Postgres would add a testcontainer
// overhead with no ratio benefit.
func BenchmarkResourceList_Paired(b *testing.B) {
	rt := benchOpenSQLite(b)
	gormRepo := store.NewResourceRepository(rt.GormDB())
	sqlcRepo := store.NewResourceRepositorySQLC(rt)
	internaltest.SeedPairedBenchFixture(b, internaltest.SeedRepos{
		Tags:      store.NewTagsRepository(rt.GormDB()),
		Channels:  store.NewNotificationChannelRepository(rt.GormDB()),
		Resources: gormRepo,
	}, internaltest.FixtureConfig{
		// Smaller than the spec default (1000×50×20) for SQLite bench
		// turnaround. Comment in retro if the ratio is too noisy.
		NumResources: 300,
		NumTags:      30,
		NumChannels:  10,
	})
	ctx := context.Background()

	internaltest.RunPairedBench(b, "BenchmarkResourceList_Paired",
		func() { _, _ = gormRepo.List(ctx, 100, 0) },
		func() { _, _ = sqlcRepo.List(ctx, 100, 0) },
	)
}
