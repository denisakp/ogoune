package store_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// TestResourceRepository_M2M_Tags_Transitions exercises SC-006 for the Tags
// side: 0→N, N→M (overlapping), N→0. Runs on both dialects via ForEachDialect
// against the sqlc impl (PR2 of US1 wired Tags M2M via WithTx).
func TestResourceRepository_M2M_Tags_Transitions(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewResourceRepositorySQLC(fx.Runtime)
		tagsRepo := store.NewTagsRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		// Seed five tags so we can exercise transitions.
		tagIDs := make([]string, 5)
		for i := 0; i < 5; i++ {
			t.Helper()
			tg := &domain.Tags{
				Base: domain.Base{ID: "", CreatedAt: time.Now()},
				Name: "tag-" + string(rune('a'+i)),
			}
			require.NoError(t, tagsRepo.Create(ctx, tg))
			tagIDs[i] = tg.ID
		}

		t.Run("0_to_N", func(t *testing.T) {
			res := &domain.Resource{
				Base:     domain.Base{ID: "m2m-0-to-n", CreatedAt: time.Now()},
				Name:     "0→N",
				Type:     domain.ResourceHTTP,
				Target:   "https://example.com",
				IsActive: true,
				Tags: []*domain.Tags{
					{Base: domain.Base{ID: tagIDs[0]}},
					{Base: domain.Base{ID: tagIDs[1]}},
				},
			}
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assertTagIDs(t, loaded.Tags, tagIDs[0], tagIDs[1])
		})

		t.Run("N_to_M_overlap", func(t *testing.T) {
			res := &domain.Resource{
				Base:     domain.Base{ID: "m2m-n-to-m", CreatedAt: time.Now()},
				Name:     "N→M",
				Type:     domain.ResourceHTTP,
				Target:   "https://example.com",
				IsActive: true,
				Tags: []*domain.Tags{
					{Base: domain.Base{ID: tagIDs[0]}},
					{Base: domain.Base{ID: tagIDs[1]}},
					{Base: domain.Base{ID: tagIDs[2]}},
				},
			}
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			// Update to a partly-overlapping set: keep tag[1], drop 0+2, add 3+4.
			res.Tags = []*domain.Tags{
				{Base: domain.Base{ID: tagIDs[1]}},
				{Base: domain.Base{ID: tagIDs[3]}},
				{Base: domain.Base{ID: tagIDs[4]}},
			}
			require.NoError(t, repo.Update(ctx, res))

			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assertTagIDs(t, loaded.Tags, tagIDs[1], tagIDs[3], tagIDs[4])
		})

		t.Run("N_to_0", func(t *testing.T) {
			res := &domain.Resource{
				Base:     domain.Base{ID: "m2m-n-to-0", CreatedAt: time.Now()},
				Name:     "N→0",
				Type:     domain.ResourceHTTP,
				Target:   "https://example.com",
				IsActive: true,
				Tags: []*domain.Tags{
					{Base: domain.Base{ID: tagIDs[0]}},
					{Base: domain.Base{ID: tagIDs[1]}},
				},
			}
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)

			res.Tags = nil
			require.NoError(t, repo.Update(ctx, res))

			loaded, err := repo.FindByID(ctx, res.ID)
			require.NoError(t, err)
			assert.Empty(t, loaded.Tags)
		})
	})
}

// TestResourceRepository_FindByTag_SqlcContract exercises the JOIN path added
// in PR2 of US1, on both dialects.
func TestResourceRepository_FindByTag_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewResourceRepositorySQLC(fx.Runtime)
		tagsRepo := store.NewTagsRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		tag := &domain.Tags{
			Base: domain.Base{ID: "", CreatedAt: time.Now()},
			Name: "shared-find-by-tag",
		}
		require.NoError(t, tagsRepo.Create(ctx, tag))

		// Two tagged resources + one untagged.
		for i, id := range []string{"fbt-1", "fbt-2"} {
			res := &domain.Resource{
				Base:     domain.Base{ID: id, CreatedAt: time.Now().Add(time.Duration(-i) * time.Minute)},
				Name:     id,
				Type:     domain.ResourceHTTP,
				Target:   "https://example.com",
				IsActive: true,
				Tags:     []*domain.Tags{{Base: domain.Base{ID: tag.ID}}},
			}
			_, err := repo.Create(ctx, res)
			require.NoError(t, err)
		}
		other := &domain.Resource{
			Base:     domain.Base{ID: "fbt-other", CreatedAt: time.Now()},
			Name:     "other",
			Type:     domain.ResourceHTTP,
			Target:   "https://example.com",
			IsActive: true,
		}
		_, err := repo.Create(ctx, other)
		require.NoError(t, err)

		got, err := repo.FindByTag(ctx, "shared-find-by-tag", 10, 0)
		require.NoError(t, err)
		ids := make([]string, len(got))
		for i, r := range got {
			ids[i] = r.ID
		}
		sort.Strings(ids)
		assert.Equal(t, []string{"fbt-1", "fbt-2"}, ids)
	})
}

func assertTagIDs(t *testing.T, tags []*domain.Tags, want ...string) {
	t.Helper()
	got := make([]string, len(tags))
	for i, tg := range tags {
		got[i] = tg.ID
	}
	sort.Strings(got)
	sort.Strings(want)
	assert.Equal(t, want, got)
}
