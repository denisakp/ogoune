package internaltest

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/pkg/crypto"
)

// Default fixture key used by SeedPairedBenchFixture. Tests can override
// APP_SECRET_KEY before calling the seeder if they need a different value.
const benchFixtureKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// FixtureConfig parameterises SeedPairedBenchFixture. Defaults match the
// spec (1000 resources × 50 tags × 20 channels, seed 42, 512 MB cap).
type FixtureConfig struct {
	NumResources int
	NumTags      int
	NumChannels  int
	Seed         int64
	MemCapBytes  int64
}

// Defaults applies the spec defaults to any zero-valued field.
func (c FixtureConfig) Defaults() FixtureConfig {
	if c.NumResources == 0 {
		c.NumResources = 1000
	}
	if c.NumTags == 0 {
		c.NumTags = 50
	}
	if c.NumChannels == 0 {
		c.NumChannels = 20
	}
	if c.Seed == 0 {
		c.Seed = 42
	}
	if c.MemCapBytes == 0 {
		c.MemCapBytes = 512 << 20
	}
	return c
}

// Fixture is the result of SeedPairedBenchFixture — IDs of seeded rows in
// deterministic order, plus diagnostics (seed wall-clock + heap delta).
type Fixture struct {
	Config       FixtureConfig
	ResourceIDs  []string
	TagIDs       []string
	ChannelIDs   []string
	SeedDuration time.Duration
	AllocDelta   int64
}

// SeedRepos is the minimal set of repositories SeedPairedBenchFixture needs
// to insert its rows. Caller constructs them (typically all GORM-backed) and
// hands them in — keeps internaltest free of an import cycle with store.
type SeedRepos struct {
	Tags      port.TagsRepository
	Channels  port.NotificationChannelRepository
	Resources port.ResourceRepository
}

// SeedPairedBenchFixture builds a deterministic dataset for paired benches.
// Caller provides the repos (typically GORM-backed; encryption hooks fire
// on channel inserts). Fails the test if heap delta exceeds MemCapBytes.
//
// Spec 049 §FR-005 / contracts/seedfixture.md.
func SeedPairedBenchFixture(tb testing.TB, repos SeedRepos, cfg FixtureConfig) *Fixture {
	tb.Helper()
	cfg = cfg.Defaults()

	tb.Setenv("APP_SECRET_KEY", benchFixtureKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})

	runtime.GC()
	var msBefore runtime.MemStats
	runtime.ReadMemStats(&msBefore)
	start := time.Now()

	prng := rand.New(rand.NewSource(cfg.Seed))
	ctx := context.Background()
	tagsRepo := repos.Tags
	chRepo := repos.Channels
	resRepo := repos.Resources

	fx := &Fixture{Config: cfg}

	// Tags.
	fx.TagIDs = make([]string, cfg.NumTags)
	for i := 0; i < cfg.NumTags; i++ {
		id := fmt.Sprintf("bench-tag-%04d", i)
		if err := tagsRepo.Create(ctx, &domain.Tags{
			Base: domain.Base{ID: id, CreatedAt: time.Now()},
			Name: id,
		}); err != nil {
			tb.Fatalf("seed tag %s: %v", id, err)
		}
		fx.TagIDs[i] = id
	}

	// Channels.
	fx.ChannelIDs = make([]string, cfg.NumChannels)
	for i := 0; i < cfg.NumChannels; i++ {
		id := fmt.Sprintf("bench-ch-%04d", i)
		if err := chRepo.Create(ctx, &domain.NotificationChannel{
			Base:   domain.Base{ID: id, CreatedAt: time.Now()},
			Name:   id,
			Type:   domain.NotificationChannelTypeSlack,
			Config: []byte(`{"webhook":"https://example.com"}`),
		}); err != nil {
			tb.Fatalf("seed channel %s: %v", id, err)
		}
		fx.ChannelIDs[i] = id
	}

	// Resources — random subset of 5 tags + 3 channels each (deterministic).
	fx.ResourceIDs = make([]string, cfg.NumResources)
	for i := 0; i < cfg.NumResources; i++ {
		id := fmt.Sprintf("bench-res-%05d", i)
		tagsSubset := subset(prng, fx.TagIDs, 5)
		chSubset := subset(prng, fx.ChannelIDs, 3)
		res := &domain.Resource{
			Base:                 domain.Base{ID: id, CreatedAt: time.Now()},
			Name:                 id,
			Type:                 domain.ResourceHTTP,
			Target:               "https://example.com",
			IsActive:             true,
			Tags:                 stubTags(tagsSubset),
			NotificationChannels: stubChannels(chSubset),
		}
		if _, err := resRepo.Create(ctx, res); err != nil {
			tb.Fatalf("seed resource %s: %v", id, err)
		}
		fx.ResourceIDs[i] = id
	}

	fx.SeedDuration = time.Since(start)

	runtime.GC()
	var msAfter runtime.MemStats
	runtime.ReadMemStats(&msAfter)
	fx.AllocDelta = int64(msAfter.Alloc) - int64(msBefore.Alloc)
	if fx.AllocDelta > cfg.MemCapBytes {
		tb.Fatalf("fixture exceeded %d MB cap (got %d MB) — reduce NumResources or raise MemCapBytes",
			cfg.MemCapBytes>>20, fx.AllocDelta>>20)
	}
	return fx
}

func subset(prng *rand.Rand, ids []string, k int) []string {
	if k >= len(ids) {
		out := make([]string, len(ids))
		copy(out, ids)
		return out
	}
	picked := make([]string, k)
	used := make(map[int]struct{}, k)
	for i := 0; i < k; i++ {
		for {
			j := prng.Intn(len(ids))
			if _, ok := used[j]; ok {
				continue
			}
			used[j] = struct{}{}
			picked[i] = ids[j]
			break
		}
	}
	return picked
}

func stubTags(ids []string) []*domain.Tags {
	out := make([]*domain.Tags, len(ids))
	for i, id := range ids {
		out[i] = &domain.Tags{Base: domain.Base{ID: id}}
	}
	return out
}

func stubChannels(ids []string) []*domain.NotificationChannel {
	out := make([]*domain.NotificationChannel, len(ids))
	for i, id := range ids {
		out[i] = &domain.NotificationChannel{Base: domain.Base{ID: id}}
	}
	return out
}
